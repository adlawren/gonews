package db

import (
	"database/sql"
	"fmt"
	"gonews/config"
	"gonews/feed"
	"gonews/timestamp"
	"gonews/user"
	"os"
	"reflect"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/pressly/goose"
)

var ErrModelNotFound = errors.New("no matching model was found")

// DB contains the methods needed to store and read data from the underlying
// database
type DB interface {
	Ping() error
	Migrate(string) error
	All(interface{}) error
	FindAll(interface{}, map[string]interface{}) error
	Find(interface{}, map[string]interface{}) error
	Save(interface{}) error
	Timestamp() (*time.Time, error)
	SaveTimestamp(*time.Time) error
	Users() ([]*user.User, error)
	MatchingUser(*user.User) (*user.User, error)
	SaveUser(*user.User) error
	Feeds() ([]*feed.Feed, error)
	MatchingFeed(*feed.Feed) (*feed.Feed, error)
	SaveFeed(*feed.Feed) error
	Tags() ([]*feed.Tag, error)
	MatchingTag(*feed.Tag) (*feed.Tag, error)
	SaveTag(*feed.Tag) error
	Items() ([]*feed.Item, error)
	Item(id uint) (*feed.Item, error)
	ItemsFromFeed(*feed.Feed) ([]*feed.Item, error)
	ItemsFromTag(*feed.Tag) ([]*feed.Item, error)
	MatchingItem(*feed.Item) (*feed.Item, error)
	SaveItem(*feed.Item) error
	Close() error
}

// New creates a struct which supports the operations in the DB interface
func New(cfg *config.DBConfig) (DB, error) {
	db, err := sql.Open("sqlite3", cfg.DSN)
	return &sqlDB{db: db}, errors.Wrap(err, "failed to open DB")
}

type sqlDB struct {
	db *sql.DB
}

func (sdb *sqlDB) Ping() error {
	return errors.Wrap(sdb.db.Ping(), "failed to ping DB")
}

func (sdb *sqlDB) Migrate(migrationsDir string) error {
	_, err := os.Stat(migrationsDir)
	if err != nil {
		return errors.Wrap(err, "failed to stat migrations directory")
	}

	err = goose.SetDialect("sqlite3")
	if err != nil {
		return errors.Wrap(err, "failed to set goose DB driver")
	}

	err = goose.Up(sdb.db, migrationsDir)
	return errors.Wrap(err, "migrations failed")
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

func (sdb *sqlDB) FindAll(ptr interface{}, attributes map[string]interface{}) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return errors.New("Invalid argument; pointer to slice of pointers to structs is required")
	}

	val := reflect.Indirect(reflect.ValueOf(ptr))
	if val.Type().Kind() != reflect.Slice {
		return errors.New("Invalid argument; pointer to slice of pointers to structs is required")
	}

	if val.Type().Elem().Kind() != reflect.Ptr {
		return errors.New("Invalid argument; pointer to slice of pointers to structs is required")
	}

	if val.Type().Elem().Elem().Kind() != reflect.Struct {
		return errors.New("Invalid argument; pointer to slice of pointers to structs is required")
	}

	modelName := val.Type().Elem().Elem().Name()
	tableName := fmt.Sprintf("%ss", toSnake(modelName))
	paramStrings := []string{}
	for attributeName := range attributes {
		paramStrings = append(paramStrings, fmt.Sprintf("%s=?", attributeName))
	}
	query := fmt.Sprintf("select * from %s where %s;", tableName, strings.Join(paramStrings, ","))

	stmt, err := sdb.db.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	attributeValues := []interface{}{}
	for _, attributeValue := range attributes {
		attributeValues = append(attributeValues, attributeValue)
	}

	rows, err := stmt.Query(attributeValues...)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

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

func (sdb *sqlDB) Find(ptr interface{}, attributes map[string]interface{}) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return errors.New("Invalid argument; pointer to struct is required")
	}

	val := reflect.Indirect(reflect.ValueOf(ptr))
	if val.Type().Kind() != reflect.Struct {
		return errors.New("Invalid argument; pointer to struct is required")
	}

	fieldPointers := []interface{}{}
	for idx := 0; idx < val.NumField(); idx++ {
		fieldPointers = append(fieldPointers, val.Field(idx).Addr().Interface())
	}

	modelName := val.Type().Name()
	tableName := fmt.Sprintf("%ss", toSnake(modelName))
	paramStrings := []string{}
	for attributeName := range attributes {
		paramStrings = append(paramStrings, fmt.Sprintf("%s=?", attributeName))
	}
	query := fmt.Sprintf("select * from %s where %s;", tableName, strings.Join(paramStrings, ","))

	stmt, err := sdb.db.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	attributeValues := []interface{}{}
	for _, attributeValue := range attributes {
		attributeValues = append(attributeValues, attributeValue)
	}

	rows, err := stmt.Query(attributeValues...)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute prepared statement")
	}

	if !rows.Next() {
		return ErrModelNotFound
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

func (sdb *sqlDB) All(ptr interface{}) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return errors.New("Invalid argument; pointer to slice of pointers to structs is required")
	}

	val := reflect.Indirect(reflect.ValueOf(ptr))
	if val.Type().Kind() != reflect.Slice {
		return errors.New("Invalid argument; pointer to slice of pointers to structs is required")
	}

	if val.Type().Elem().Kind() != reflect.Ptr {
		return errors.New("Invalid argument; pointer to slice of pointers to structs is required")
	}

	if val.Type().Elem().Elem().Kind() != reflect.Struct {
		return errors.New("Invalid argument; pointer to slice of pointers to structs is required")
	}

	modelName := val.Type().Elem().Elem().Name()
	tableName := fmt.Sprintf("%ss", toSnake(modelName))
	query := fmt.Sprintf("select * from %s;", tableName)

	rows, err := sdb.db.Query(query)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

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

func (sdb *sqlDB) insertModel(ptr interface{}) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return errors.New("Invalid argument; pointer to struct is required")
	}

	val := reflect.Indirect(reflect.ValueOf(ptr))
	if val.Type().Kind() != reflect.Struct {
		return errors.New("Invalid argument; pointer to struct is required")
	}

	snakeFieldNames := []string{}
	fieldValues := []interface{}{}
	for idx := 0; idx < val.NumField(); idx++ {
		if val.Type().Field(idx).Name == "ID" {
			continue
		}
		if val.Type().Field(idx).Name == "CreatedAt" {
			continue
		}

		snakeFieldNames = append(snakeFieldNames, toSnake(val.Type().Field(idx).Name))
		fieldValues = append(fieldValues, val.Field(idx).Interface())
	}

	// If a model has a CreatedAt field, set it to the current time
	var zeroValue reflect.Value
	if val.FieldByName("CreatedAt") != zeroValue {
		snakeFieldNames = append(snakeFieldNames, "created_at")
		fieldValues = append(fieldValues, reflect.ValueOf(time.Now()).Interface())
	}

	modelName := val.Type().Name()
	tableName := fmt.Sprintf("%ss", toSnake(modelName))
	paramStrings := []string{}
	for idx := 0; idx < len(snakeFieldNames); idx++ {
		paramStrings = append(paramStrings, "?")
	}
	query := fmt.Sprintf("insert into %s (%s) values (%s);", tableName, strings.Join(snakeFieldNames, ","), strings.Join(paramStrings, ","))

	stmt, err := sdb.db.Prepare(query)
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

	val.FieldByName("ID").Set(reflect.ValueOf(uint(id)))

	return nil
}

func (sdb *sqlDB) updateModel(ptr interface{}) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return errors.New("Invalid argument; pointer to struct is required")
	}

	val := reflect.Indirect(reflect.ValueOf(ptr))
	if val.Type().Kind() != reflect.Struct {
		return errors.New("Invalid argument; pointer to struct is required")
	}

	fieldNames := []string{}
	fieldValues := []interface{}{}
	for idx := 0; idx < val.NumField(); idx++ {
		if val.Type().Field(idx).Name == "ID" {
			continue
		}
		if val.Type().Field(idx).Name == "CreatedAt" {
			continue
		}

		fieldNames = append(fieldNames, val.Type().Field(idx).Name)
		fieldValues = append(fieldValues, val.Field(idx).Interface())
	}
	fieldValues = append(fieldValues, val.FieldByName("ID").Interface())

	modelName := val.Type().Name()
	tableName := fmt.Sprintf("%ss", toSnake(modelName))
	paramStrings := []string{}
	for _, fieldName := range fieldNames {
		paramStrings = append(paramStrings, fmt.Sprintf("%s=?", toSnake(fieldName)))
	}
	query := fmt.Sprintf("update %s set %s where id=?;", tableName, strings.Join(paramStrings, ","))

	stmt, err := sdb.db.Prepare(query)
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

func (sdb *sqlDB) Save(ptr interface{}) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return errors.New("Invalid argument; pointer to struct is required")
	}

	val := reflect.Indirect(reflect.ValueOf(ptr))
	if val.Type().Kind() != reflect.Struct {
		return errors.New("Invalid argument; pointer to struct is required")
	}

	modelName := val.Type().Name()
	tableName := fmt.Sprintf("%ss", toSnake(modelName))
	query := fmt.Sprintf("select count(*) from %s where id=?;", tableName)

	stmt, err := sdb.db.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	idValue := val.FieldByName("ID").Interface()

	rows, err := stmt.Query(idValue)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute prepared statement")
	}

	var count int
	if !rows.Next() {
		return errors.New("cursor is empty")
	}
	err = rows.Scan(&count)
	if err != nil {
		return errors.Wrap(err, "failed to scan count")
	}

	err = rows.Err()
	if err != nil {
		return errors.Wrap(err, "cursor error")
	}

	err = rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close cursor")
	}

	if count != 0 {
		err = sdb.updateModel(ptr)
	} else {
		err = sdb.insertModel(ptr)
	}

	return errors.Wrap(err, "failed to save model")
}

func scanTimestamp(rows *sql.Rows, ts *timestamp.Timestamp) error {
	return rows.Scan(&ts.ID, &ts.T)
}

func (sdb *sqlDB) Timestamp() (*time.Time, error) {
	rows, err := sdb.db.Query("select id, t from timestamps;")
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	var ts timestamp.Timestamp
	if rows.Next() {
		err = scanTimestamp(rows, &ts)
	}

	err = rows.Err()
	if err != nil {
		return &ts.T, errors.Wrap(err, "cursor error")
	}

	return &ts.T, errors.Wrap(err, "failed to scan timestamp")
}

func (sdb *sqlDB) insertTimestamp(ts *timestamp.Timestamp) error {
	stmt, err := sdb.db.Prepare("insert into timestamps (t) values (?);")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	res, err := stmt.Exec(ts.T)
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

	ts.ID = uint(id)

	return nil
}

func (sdb *sqlDB) updateTimestamp(ts *timestamp.Timestamp) error {
	stmt, err := sdb.db.Prepare("update timestamps set t=? where id=?;")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	res, err := stmt.Exec(ts.T, ts.ID)
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

func (sdb *sqlDB) SaveTimestamp(t *time.Time) error {
	rows, err := sdb.db.Query("select count(*) from timestamps;")
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	var count int
	if !rows.Next() {
		return errors.New("no timestamps found")
	}
	err = rows.Scan(&count)
	if err != nil {
		return errors.Wrap(err, "failed to scan count")
	}

	err = rows.Err()
	if err != nil {
		return errors.Wrap(err, "cursor error")
	}

	if count > 1 {
		err = errors.New("multiple timestamps in DB")
	}

	err = rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close cursor")
	}

	rows, err = sdb.db.Query("select id, t from timestamps limit 1;")
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	var ts timestamp.Timestamp
	if rows.Next() {
		err = scanTimestamp(rows, &ts)
	}
	if err != nil {
		return errors.Wrap(err, "failed to scan timestamp")
	}

	err = rows.Err()
	if err != nil {
		return errors.Wrap(err, "cursor error")
	}

	err = rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close cursor")
	}

	ts.T = *t
	if count == 0 {
		err = sdb.insertTimestamp(&ts)
	} else {
		err = sdb.updateTimestamp(&ts)
	}

	return errors.Wrap(err, "failed to save timestamp")
}

func (sdb *sqlDB) Users() ([]*user.User, error) {
	var users []*user.User
	err := sdb.All(&users)
	return users, errors.Wrap(err, "failed to get users")
}

func (sdb *sqlDB) MatchingUser(u *user.User) (*user.User, error) {
	var user user.User
	err := sdb.Find(&user, map[string]interface{}{"username": u.Username})
	if err == ErrModelNotFound {
		return nil, nil
	}
	return &user, errors.Wrap(err, "failed to get matching user")
}

func (sdb *sqlDB) SaveUser(u *user.User) error {
	err := sdb.Save(u)
	return errors.Wrap(err, "failed to save user")
}

func (sdb *sqlDB) Feeds() ([]*feed.Feed, error) {
	var feeds []*feed.Feed
	err := sdb.All(&feeds)
	return feeds, errors.Wrap(err, "failed to get all feeds")
}

func (sdb *sqlDB) FeedsFromTag(t *feed.Tag) ([]*feed.Feed, error) {
	var feeds []*feed.Feed

	var tag feed.Tag
	err := sdb.Find(&tag, map[string]interface{}{"name": tag.Name})
	if err != nil {
		return feeds, errors.Wrap(err, "failed to find tag")
	}

	err = sdb.FindAll(&feeds, map[string]interface{}{"tag_id": tag.ID})
	return feeds, errors.Wrap(err, "failed to get feeds")
}

func (sdb *sqlDB) MatchingFeed(f *feed.Feed) (*feed.Feed, error) {
	var feed feed.Feed
	err := sdb.Find(&feed, map[string]interface{}{"url": f.URL})
	if err == ErrModelNotFound {
		return nil, nil
	}
	return &feed, errors.Wrap(err, "failed to get matching feed")
}

func (sdb *sqlDB) SaveFeed(f *feed.Feed) error {
	err := sdb.Save(f)
	return errors.Wrap(err, "failed to save feed")
}

func (sdb *sqlDB) Tags() ([]*feed.Tag, error) {
	var tags []*feed.Tag
	err := sdb.All(&tags)
	return tags, errors.Wrap(err, "failed to get all tags")
}

func (sdb *sqlDB) MatchingTag(t *feed.Tag) (*feed.Tag, error) {
	var tag feed.Tag
	err := sdb.Find(&tag, map[string]interface{}{"name": t.Name})
	if err == ErrModelNotFound {
		return nil, nil
	}
	return &tag, errors.Wrap(err, "failed to get matching tag")
}

func (sdb *sqlDB) SaveTag(t *feed.Tag) error {
	err := sdb.Save(t)
	return errors.Wrap(err, "failed to save tag")
}

func (sdb *sqlDB) Items() ([]*feed.Item, error) {
	var items []*feed.Item
	err := sdb.All(&items)
	return items, errors.Wrap(err, "failed to get all items")
}

func (sdb *sqlDB) MatchingItem(i *feed.Item) (*feed.Item, error) {
	var item feed.Item
	err := sdb.Find(&item, map[string]interface{}{"link": i.Link})
	if err == ErrModelNotFound {
		return nil, nil
	}
	return &item, errors.Wrap(err, "failed to get matching item")
}

func (sdb *sqlDB) Item(id uint) (*feed.Item, error) {
	var item feed.Item
	err := sdb.Find(&item, map[string]interface{}{"id": id})
	return &item, errors.Wrap(err, "failed to find item")
}

func (sdb *sqlDB) ItemsFromFeed(f *feed.Feed) ([]*feed.Item, error) {
	var items []*feed.Item
	err := sdb.FindAll(&items, map[string]interface{}{"feed_id": f.ID})
	return items, errors.Wrap(err, "failed to find items")
}

func (sdb *sqlDB) ItemsFromTag(t *feed.Tag) ([]*feed.Item, error) {
	var items []*feed.Item

	var tag feed.Tag
	err := sdb.Find(&tag, map[string]interface{}{"name": t.Name})
	if err != nil {
		return items, errors.Wrap(err, "failed to find tag")
	}

	var feeds []*feed.Feed
	err = sdb.FindAll(&feeds, map[string]interface{}{"tag_id": tag.ID})
	if err != nil {
		return items, errors.Wrap(err, "failed to find feeds")
	}

	for _, f := range feeds {
		var nextItems []*feed.Item
		err = sdb.FindAll(&items, map[string]interface{}{"feed_id": f.ID})
		if err != nil {
			return items, errors.Wrap(err, "failed to get items")
		}
		items = append(items, nextItems...)
	}

	return items, nil
}

func (sdb *sqlDB) SaveItem(i *feed.Item) error {
	err := sdb.Save(i)
	return errors.Wrap(err, "failed to save item")
}

func (sdb *sqlDB) Close() error {
	err := sdb.db.Close()
	return errors.Wrap(err, "failed to close DB")
}
