package query

import (
	"database/sql"
	"fmt"
	"gonews/db/orm/query/clause"
	"reflect"
	"strings"
	"time"
)

var ErrMissingIdField = fmt.Errorf("struct must contain ID field")
var ErrModelNotFound = fmt.Errorf("no matching model was found")

var ErrInvalidModelArg = fmt.Errorf("invalid argument; pointer to struct is required")
var ErrInvalidModelsArg = fmt.Errorf("invalid argument; pointer to slice of pointers to structs is required")

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

func modelTypeHasID(model interface{}) bool {
	_, found := modelType(model).FieldByName("ID")
	return found
}

func modelsTypeHasID(models interface{}) bool {
	_, found := modelsType(models).FieldByName("ID")
	return found
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

// Query contains the methods needed to execute a SQL query in a given database/transaction
type Query interface {
	Exec(*sql.DB) error
	ExecTx(*sql.Tx) error
}

type query struct {
	str  string
	args []interface{}
}

func (q *query) add(clause *clause.Clause) {
	q.str = fmt.Sprintf("%s %s", q.str, clause.Text())
	q.args = append(q.args, clause.Args()...)
}

func (q *query) addAll(clauses ...*clause.Clause) {
	for _, clause := range clauses {
		q.add(clause)
	}
}

func exec(q Query, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to create DB transaction: %w", err)
	}
	defer tx.Rollback()

	err = q.ExecTx(tx)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit DB transaction: %w", err)
	}

	return nil
}

type deleteQuery struct {
	query
	models interface{}
}

func (q *deleteQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *deleteQuery) ExecTx(tx *sql.Tx) error {
	modelsVal := reflect.Indirect(reflect.ValueOf(q.models))

	var ids []interface{}
	var paramStrings []string
	for i := 0; i < modelsVal.Len(); i++ {
		modelVal := reflect.Indirect(modelsVal.Index(i))
		modelId := modelVal.FieldByName("ID").Uint()
		ids = append(ids, uint(modelId))
		paramStrings = append(paramStrings, "?")
	}

	inClause := clause.New(fmt.Sprintf("in (%s)", strings.Join(paramStrings, ",")), ids...)

	q.addAll(clause.New("where id"), inClause)

	stmt, err := tx.Prepare(q.str)
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	res, err := stmt.Exec(ids...)
	if err != nil {
		return fmt.Errorf("failed to execute prepared statement: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected row count: %w", err)
	}
	if count != int64(len(ids)) {
		return fmt.Errorf("expected all models to be deleted")
	}

	return nil
}

type selectCountQuery struct {
	query
	result *int
}

func (q *selectCountQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *selectCountQuery) ExecTx(tx *sql.Tx) error {
	stmt, err := tx.Prepare(q.str)
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("failed to prepare statement")
	}

	rows, err := stmt.Query(q.args...)
	defer rows.Close()
	if err != nil {
		return fmt.Errorf("failed to execute query")
	}

	if !rows.Next() {
		return fmt.Errorf("not result returned")
	}

	err = rows.Scan(q.result)
	if err != nil {
		return fmt.Errorf("failed to scan result")
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	return nil
}

type selectQuery struct {
	query
	results interface{}
}

func (q *selectQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *selectQuery) ExecTx(tx *sql.Tx) error {
	stmt, err := tx.Prepare(q.str)
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	rows, err := stmt.Query(q.args...)
	defer rows.Close()
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	val := reflect.Indirect(reflect.ValueOf(q.results))
	slicValue := reflect.Indirect(reflect.New(reflect.SliceOf(val.Type().Elem())))
	for rows.Next() {
		modelValue := reflect.Indirect(reflect.New(val.Type().Elem().Elem()))
		fieldPointers := []interface{}{}
		for idx := 0; idx < modelValue.NumField(); idx++ {
			fieldPointers = append(fieldPointers, modelValue.Field(idx).Addr().Interface())
		}

		err = rows.Scan(fieldPointers...)
		if err != nil {
			return fmt.Errorf("failed to scan model: %w", err)
		}

		slicValue = reflect.Append(slicValue, modelValue.Addr())
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	val.Set(slicValue)

	return nil
}

type selectOneQuery struct {
	query
	result interface{}
}

func (q *selectOneQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *selectOneQuery) ExecTx(tx *sql.Tx) error {
	stmt, err := tx.Prepare(q.str)
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	rows, err := stmt.Query(q.args...)
	defer rows.Close()
	if err != nil {
		return fmt.Errorf("failed to execute prepared statement: %w", err)
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
		return fmt.Errorf("failed to scan model: %w", err)
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	return nil
}

type insertQuery struct {
	query
	model interface{}
}

func (q *insertQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *insertQuery) ExecTx(tx *sql.Tx) error {
	modelVal := reflect.Indirect(reflect.ValueOf(q.model))

	snakeFieldNames := []string{}
	fieldValues := []interface{}{}
	for idx := 0; idx < modelVal.NumField(); idx++ {
		// Caller shouldn't be modifying ID
		if modelVal.Type().Field(idx).Name == "ID" {
			continue
		}

		// CreatedAt will be set below if present
		if modelVal.Type().Field(idx).Name == "CreatedAt" {
			continue
		}

		// UpdatedAt will be set below if present
		if modelVal.Type().Field(idx).Name == "UpdatedAt" {
			continue
		}

		snakeFieldNames = append(snakeFieldNames, toSnake(modelVal.Type().Field(idx).Name))
		fieldValues = append(fieldValues, modelVal.Field(idx).Interface())
	}

	// For CreatedAt/UpdatedAt
	now := time.Now()

	// If a model has a CreatedAt field, set it to the current time
	var zeroValue reflect.Value
	if modelVal.FieldByName("CreatedAt") != zeroValue {
		snakeFieldNames = append(snakeFieldNames, "created_at")
		fieldValues = append(fieldValues, reflect.ValueOf(now).Interface())
		modelVal.FieldByName("CreatedAt").Set(reflect.ValueOf(now))
	}

	// If a model has an UpdatedAt field, set it to the current time
	if modelVal.FieldByName("UpdatedAt") != zeroValue {
		snakeFieldNames = append(snakeFieldNames, "updated_at")
		fieldValues = append(fieldValues, reflect.ValueOf(now).Interface())
		modelVal.FieldByName("UpdatedAt").Set(reflect.ValueOf(now))
	}

	modelName := modelVal.Type().Name()
	tableName := fmt.Sprintf("%ss", toSnake(modelName))
	paramStrings := []string{}
	for idx := 0; idx < len(snakeFieldNames); idx++ {
		paramStrings = append(paramStrings, "?")
	}
	query := fmt.Sprintf("insert into %s (%s) values (%s)", tableName, strings.Join(snakeFieldNames, ","), strings.Join(paramStrings, ","))

	stmt, err := tx.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	res, err := stmt.Exec(fieldValues...)
	if err != nil {
		return fmt.Errorf("failed to execute prepared statement: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected row count: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("expected one row to be affected")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last inserted id: %w", err)
	}

	modelVal.FieldByName("ID").Set(reflect.ValueOf(uint(id)))

	return nil
}

type updateQuery struct {
	query
	model interface{}
}

func (q *updateQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *updateQuery) ExecTx(tx *sql.Tx) error {
	modelVal := reflect.Indirect(reflect.ValueOf(q.model))

	fieldNames := []string{}
	fieldValues := []interface{}{}
	for idx := 0; idx < modelVal.NumField(); idx++ {
		// Caller shouldn't be modifying ID
		if modelVal.Type().Field(idx).Name == "ID" {
			continue
		}

		// CreatedAt is set in a create query; shouldn't be modified by caller
		if modelVal.Type().Field(idx).Name == "CreatedAt" {
			continue
		}

		// UpdatedAt will be set below if present
		if modelVal.Type().Field(idx).Name == "UpdatedAt" {
			continue
		}

		fieldNames = append(fieldNames, modelVal.Type().Field(idx).Name)
		fieldValues = append(fieldValues, modelVal.Field(idx).Interface())
	}

	// If a model has an UpdatedAt field, set it to the current time
	var zeroValue reflect.Value
	if modelVal.FieldByName("UpdatedAt") != zeroValue {
		fieldNames = append(fieldNames, "UpdatedAt")

		now := time.Now()
		fieldValues = append(fieldValues, reflect.ValueOf(now).Interface())
		modelVal.FieldByName("UpdatedAt").Set(reflect.ValueOf(now))
	}

	fieldValues = append(fieldValues, modelVal.FieldByName("ID").Interface())

	modelName := modelVal.Type().Name()
	tableName := fmt.Sprintf("%ss", toSnake(modelName))
	paramStrings := []string{}
	for _, fieldName := range fieldNames {
		paramStrings = append(paramStrings, fmt.Sprintf("%s=?", toSnake(fieldName)))
	}
	query := fmt.Sprintf("update %s set %s where id=?", tableName, strings.Join(paramStrings, ","))

	stmt, err := tx.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	res, err := stmt.Exec(fieldValues...)
	if err != nil {
		return fmt.Errorf("failed to execute prepared statement: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected row count: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("expected one row to be affected")
	}

	return nil
}

type upsertQuery struct {
	query
	model interface{}
}

func (q *upsertQuery) Exec(db *sql.DB) error {
	return exec(q, db)
}

func (q *upsertQuery) ExecTx(tx *sql.Tx) error {
	var count int
	modelVal := reflect.Indirect(reflect.ValueOf(q.model))
	selectCountFromQuery := SelectCountFrom(modelTable(q.model), &count, clause.New("where id = ?", modelVal.FieldByName("ID").Interface()))

	err := selectCountFromQuery.ExecTx(tx)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	var query Query
	if count > 0 {
		query, err = Update(q.model)
	} else {
		query, err = Insert(q.model)
	}

	if err != nil {
		return fmt.Errorf("failed to create query: %w", err)
	}

	err = query.ExecTx(tx)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// Delete returns a delete query which deletes the given models from the appropriate table
func Delete(models interface{}) (Query, error) {
	var query deleteQuery

	if !isModels(models) {
		return &query, ErrInvalidModelsArg
	}

	if !modelsTypeHasID(models) {
		return &query, ErrMissingIdField
	}

	query.str = fmt.Sprintf("delete from %s", modelsTable(models))
	query.models = models

	return &query, nil
}

// SelectCountFrom returns a select query which fetches the number of records in the given table and assigns the result to the given reference, subject to the given query clauses
func SelectCountFrom(table string, result *int, clauses ...*clause.Clause) Query {
	var query selectCountQuery

	query.str = fmt.Sprintf("select count(*) from %s", table)
	query.result = result
	query.addAll(clauses...)

	return &query
}

// Select returns a select query which fetches the models from the appropriate table and assigns the result to the given interface, subject to the given query clauses
func Select(results interface{}, clauses ...*clause.Clause) (Query, error) {
	var query selectQuery

	if !isModels(results) {
		return &query, ErrInvalidModelsArg
	}

	if !modelsTypeHasID(results) {
		return &query, ErrMissingIdField
	}

	query.str = fmt.Sprintf("select * from %s", modelsTable(results))
	query.results = results
	query.addAll(clauses...)

	return &query, nil
}

// SelectOne returns a select query which fetches the first model from the appropriate table and assigns the result to the given interface, subject to the given query clauses
func SelectOne(result interface{}, clauses ...*clause.Clause) (Query, error) {
	var query selectOneQuery

	if !isModel(result) {
		return &query, ErrInvalidModelArg
	}

	if !modelTypeHasID(result) {
		return &query, ErrMissingIdField
	}

	query.str = fmt.Sprintf("select * from %s", modelTable(result))
	query.result = result
	query.addAll(clauses...)

	return &query, nil
}

// Insert returns an insert query which inserts the model into the appropriate table
func Insert(model interface{}) (Query, error) {
	var query insertQuery

	if !isModel(model) {
		return &query, ErrInvalidModelArg
	}

	if !modelTypeHasID(model) {
		return &query, ErrMissingIdField
	}

	query.model = model

	return &query, nil
}

// Update returns an update query which updates the model in the appropriate table
func Update(model interface{}) (Query, error) {
	var query updateQuery

	if !isModel(model) {
		return &query, ErrInvalidModelArg
	}

	if !modelTypeHasID(model) {
		return &query, ErrMissingIdField
	}

	query.model = model

	return &query, nil
}

// Upsert returns an upsert query which inserts the model into the appropriate table if it has an unspecified ID, and updates it otherwise
func Upsert(model interface{}) (Query, error) {
	var query upsertQuery

	if !isModel(model) {
		return &query, ErrInvalidModelArg
	}

	if !modelTypeHasID(model) {
		return &query, ErrMissingIdField
	}

	query.model = model

	return &query, nil
}
