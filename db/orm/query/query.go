package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var ErrModelNotFound = errors.New("no matching model was found")

var ErrInvalidModelArg = errors.New("invalid argument; pointer to struct is required")
var ErrInvalidModelsArg = errors.New("invalid argument; pointer to slice of pointers to structs is required")

func isModel(inter interface{}) bool {
	if reflect.TypeOf(inter).Kind() != reflect.Ptr {
		return false
	}

	val := reflect.Indirect(reflect.ValueOf(inter))
	if val.Type().Kind() != reflect.Struct {
		return false
	}

	return true
}

func isModels(inter interface{}) bool {
	if reflect.TypeOf(inter).Kind() != reflect.Ptr {
		return false
	}

	val := reflect.Indirect(reflect.ValueOf(inter))
	if val.Type().Kind() != reflect.Slice {
		return false
	}

	if val.Type().Elem().Kind() != reflect.Ptr {
		return false
	}

	if val.Type().Elem().Elem().Kind() != reflect.Struct {
		return false
	}

	return true
}

func modelType(model interface{}) reflect.Type {
	return reflect.Indirect(reflect.ValueOf(model)).Type()
}

func modelsType(models interface{}) reflect.Type {
	return reflect.Indirect(reflect.ValueOf(models)).Type().Elem().Elem()
}

func isUpper(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

func isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}

	return b
}

func toSnake(s string) string {
	var b strings.Builder
	for idx := range s {
		char := s[idx]
		var prevChar, nextChar byte
		if idx > 1 {
			prevChar = s[idx-1]
		}
		if idx < len(s)-2 {
			nextChar = s[idx+1]
		}
		if isUpper(char) && isLower(prevChar) {
			b.WriteRune('_')
		} else if isUpper(char) && isLower(nextChar) && b.Len() != 0 {
			b.WriteRune('_')
		}

		b.WriteByte(toLower(char))
	}

	return b.String()
}

func modelTable(model interface{}) string {
	modelType := modelType(model)
	modelName := modelType.Name()
	return fmt.Sprintf("%ss", toSnake(modelName))
}

func modelsTable(models interface{}) string {
	modelType := modelsType(models)
	modelName := modelType.Name()
	return fmt.Sprintf("%ss", toSnake(modelName))
}

// Clause represents an SQL clause in a SQL query
type Clause struct {
	str  string
	args []interface{}
}

// NewClause creates a Clause from the given arguments
func NewClause(clause string, args ...interface{}) *Clause {
	return &Clause{
		str:  clause,
		args: args,
	}
}

// Query contains the methods needed to execute a SQL query in a given database/transaction
type Query interface {
	Exec(*sql.DB) error
	ExecTx(*sql.Tx) error
}

type query struct {
	str    string
	args   []interface{}
	result interface{}
}

func (q *query) add(clause *Clause) {
	q.str = fmt.Sprintf("%s %s", q.str, clause.str)
	q.args = append(q.args, clause.args...)
}

func (q *query) addAll(clauses ...*Clause) {
	for _, clause := range clauses {
		q.add(clause)
	}
}

func exec(q Query, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to create DB transaction")
	}
	defer tx.Rollback()

	err = q.ExecTx(tx)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	return errors.Wrap(tx.Commit(), "failed to commit DB transaction")
}

type selectCountQuery struct {
	query
}

func (q *selectCountQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *selectCountQuery) ExecTx(tx *sql.Tx) error {
	stmt, err := tx.Prepare(q.str)
	defer stmt.Close()
	if err != nil {
		return errors.New("failed to prepare statement")
	}

	rows, err := stmt.Query(q.args...)
	defer rows.Close()
	if err != nil {
		return errors.New("failed to execute query")
	}

	if !rows.Next() {
		return errors.New("not result returned")
	}

	err = rows.Scan(q.result)
	if err != nil {
		return errors.New("failed to scan result")
	}

	err = rows.Err()
	if err != nil {
		return errors.Wrap(err, "cursor error")
	}

	return nil
}

type selectQuery struct {
	query
}

func (q *selectQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *selectQuery) ExecTx(tx *sql.Tx) error {
	stmt, err := tx.Prepare(q.str)
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(q.args...)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	val := reflect.Indirect(reflect.ValueOf(q.result))
	slicValue := reflect.Indirect(reflect.New(reflect.SliceOf(val.Type().Elem())))
	for rows.Next() {
		modelValue := reflect.Indirect(reflect.New(val.Type().Elem().Elem()))
		fieldPointers := []interface{}{}
		for idx := 0; idx < modelValue.NumField(); idx++ {
			fieldPointers = append(fieldPointers, modelValue.Field(idx).Addr().Interface())
		}

		err = rows.Scan(fieldPointers...)
		if err != nil {
			return errors.Wrap(err, "failed to scan model")
		}

		slicValue = reflect.Append(slicValue, modelValue.Addr())
	}

	err = rows.Err()
	if err != nil {
		return errors.Wrap(err, "cursor error")
	}

	val.Set(slicValue)

	return nil
}

type selectOneQuery struct {
	query
}

func (q *selectOneQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *selectOneQuery) ExecTx(tx *sql.Tx) error {
	stmt, err := tx.Prepare(q.str)
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(q.args...)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute prepared statement")
	}

	if !rows.Next() {
		return ErrModelNotFound
	}

	val := reflect.Indirect(reflect.ValueOf(q.result))

	fieldPointers := []interface{}{}
	for idx := 0; idx < val.NumField(); idx++ {
		fieldPointers = append(fieldPointers, val.Field(idx).Addr().Interface())
	}

	err = rows.Scan(fieldPointers...)
	if err != nil {
		return errors.Wrap(err, "failed to scan model")
	}

	err = rows.Err()
	if err != nil {
		return errors.Wrap(err, "cursor error")
	}

	return nil
}

type insertQuery struct {
	query
}

func (q *insertQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *insertQuery) ExecTx(tx *sql.Tx) error {
	resultVal := reflect.Indirect(reflect.ValueOf(q.result))

	snakeFieldNames := []string{}
	fieldValues := []interface{}{}
	for idx := 0; idx < resultVal.NumField(); idx++ {
		// Caller shouldn't be modifying ID
		if resultVal.Type().Field(idx).Name == "ID" {
			continue
		}

		// CreatedAt will be set below if present
		if resultVal.Type().Field(idx).Name == "CreatedAt" {
			continue
		}

		snakeFieldNames = append(snakeFieldNames, toSnake(resultVal.Type().Field(idx).Name))
		fieldValues = append(fieldValues, resultVal.Field(idx).Interface())
	}

	// If a model has a CreatedAt field, set it to the current time
	var zeroValue reflect.Value
	if resultVal.FieldByName("CreatedAt") != zeroValue {
		snakeFieldNames = append(snakeFieldNames, "created_at")
		fieldValues = append(fieldValues, reflect.ValueOf(time.Now()).Interface())
	}

	modelName := resultVal.Type().Name()
	tableName := fmt.Sprintf("%ss", toSnake(modelName))
	paramStrings := []string{}
	for idx := 0; idx < len(snakeFieldNames); idx++ {
		paramStrings = append(paramStrings, "?")
	}
	query := fmt.Sprintf("insert into %s (%s) values (%s)", tableName, strings.Join(snakeFieldNames, ","), strings.Join(paramStrings, ","))

	stmt, err := tx.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	res, err := stmt.Exec(fieldValues...)
	if err != nil {
		return errors.Wrap(err, "failed to execute prepared statement")
	}

	count, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get affected row count")
	}
	if count != 1 {
		return errors.New("expected one row to be affected")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "failed to get last inserted id")
	}

	resultVal.FieldByName("ID").Set(reflect.ValueOf(uint(id)))

	return nil
}

type updateQuery struct {
	query
}

func (q *updateQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *updateQuery) ExecTx(tx *sql.Tx) error {
	resultVal := reflect.Indirect(reflect.ValueOf(q.result))

	fieldNames := []string{}
	fieldValues := []interface{}{}
	for idx := 0; idx < resultVal.NumField(); idx++ {
		// Caller shouldn't be modifying ID
		if resultVal.Type().Field(idx).Name == "ID" {
			continue
		}

		// CreatedAt is set in a create query; shouldn't be modified by caller
		if resultVal.Type().Field(idx).Name == "CreatedAt" {
			continue
		}

		fieldNames = append(fieldNames, resultVal.Type().Field(idx).Name)
		fieldValues = append(fieldValues, resultVal.Field(idx).Interface())
	}
	fieldValues = append(fieldValues, resultVal.FieldByName("ID").Interface())

	modelName := resultVal.Type().Name()
	tableName := fmt.Sprintf("%ss", toSnake(modelName))
	paramStrings := []string{}
	for _, fieldName := range fieldNames {
		paramStrings = append(paramStrings, fmt.Sprintf("%s=?", toSnake(fieldName)))
	}
	query := fmt.Sprintf("update %s set %s where id=?", tableName, strings.Join(paramStrings, ","))

	stmt, err := tx.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	res, err := stmt.Exec(fieldValues...)
	if err != nil {
		return errors.Wrap(err, "failed to execute prepared statement")
	}

	count, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get affected row count")
	}
	if count != 1 {
		return errors.New("expected one row to be affected")
	}

	return nil
}

type upsertQuery struct {
	query
}

func (q *upsertQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *upsertQuery) ExecTx(tx *sql.Tx) error {
	var count int
	resultVal := reflect.Indirect(reflect.ValueOf(q.result))
	selectCountFromQuery := SelectCountFrom(modelTable(q.result), &count, NewClause("where id = ?", resultVal.FieldByName("ID").Interface()))

	err := selectCountFromQuery.ExecTx(tx)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	var query Query
	if count > 0 {
		query, err = Update(q.result)
	} else {
		query, err = Insert(q.result)
	}

	if err != nil {
		return errors.Wrap(err, "failed to create query")
	}

	err = query.ExecTx(tx)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	return nil
}

// SelectCountFrom returns a select query which fetches the number of records in the given table and assigns the result to the given reference, subject to the given query clauses
func SelectCountFrom(table string, result *int, clauses ...*Clause) Query {
	var query selectCountQuery

	query.str = fmt.Sprintf("select count(*) from %s", table)
	query.result = result
	query.addAll(clauses...)

	return &query
}

// Select returns a select query which fetches the models from the appropriate table and assigns the result to the given interface, subject to the given query clauses
func Select(result interface{}, clauses ...*Clause) (Query, error) {
	var query selectQuery

	if !isModels(result) {
		return &query, ErrInvalidModelsArg
	}

	query.str = fmt.Sprintf("select * from %s", modelsTable(result))
	query.result = result
	query.addAll(clauses...)

	return &query, nil
}

// SelectOne returns a select query which fetches the first model from the appropriate table and assigns the result to the given interface, subject to the given query clauses
func SelectOne(result interface{}, clauses ...*Clause) (Query, error) {
	var query selectOneQuery

	if !isModel(result) {
		return &query, ErrInvalidModelArg
	}

	query.str = fmt.Sprintf("select * from %s", modelTable(result))
	query.result = result
	query.addAll(clauses...)

	return &query, nil
}

// Insert returns an insert query which inserts the model into the appropriate table
func Insert(result interface{}) (Query, error) {
	var query insertQuery

	if !isModel(result) {
		return &query, ErrInvalidModelArg
	}

	query.result = result

	return &query, nil
}

// Update returns an update query which updates the model in the appropriate table
func Update(result interface{}) (Query, error) {
	var query updateQuery

	if !isModel(result) {
		return &query, ErrInvalidModelArg
	}

	query.result = result

	return &query, nil
}

// Upsert returns an upsert query which inserts the model into the appropriate table if it has an unspecified ID, and updates the model otherwise
func Upsert(result interface{}) (Query, error) {
	var query upsertQuery

	if !isModel(result) {
		return &query, ErrInvalidModelArg
	}

	query.result = result

	return &query, nil
}