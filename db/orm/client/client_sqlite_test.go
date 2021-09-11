package client

import (
	"database/sql"
	"errors"
	"fmt"
	"gonews/db/orm/query"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

type Model struct {
	ID     uint
	Bool   bool
	String string
}

type ManagedFieldsModel struct {
	ID        uint
	Bool      bool
	String    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func initDB(t *testing.T) *sql.DB {
	path := fmt.Sprintf(
		"/tmp/gonews/test/%d/db.sqlite3",
		time.Now().UnixNano())
	fmt.Printf("Initializing test DB: %s\n", path)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		goto createDB
	}

	assert.NoError(t, err)

createDB:
	err = os.MkdirAll(filepath.Dir(path), os.ModeDir)
	assert.NoError(t, err)

	_, err = os.Create(path)
	assert.NoError(t, err)

	dsn := fmt.Sprintf("file:%s", path)

	db, err := sql.Open("sqlite3", dsn)
	assert.NoError(t, err)

	createModelsTable(t, db)

	return db
}

func createModelsTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS \"models\" (\"id\" integer primary key autoincrement,\"bool\" bool,\"string\" varchar(255));")
	assert.NoError(t, err)
}

func createManagedFieldsModelsTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS \"managed_fields_models\" (\"id\" integer primary key autoincrement,\"bool\" bool,\"string\" varchar(255),\"created_at\" datetime, \"updated_at\" datetime);")
	assert.NoError(t, err)
}

func assertModelsEqual(t *testing.T, m1, m2 *Model) {
	assert.Equal(t, m1.Bool, m2.Bool)
	assert.Equal(t, m1.String, m2.String)
}

func TestAll(t *testing.T) {
	db := initDB(t)
	client := New(db)

	model1 := Model{
		Bool:   true,
		String: "abc",
	}
	model2 := Model{
		Bool:   true,
		String: "def",
	}

	err := client.Save(&model1)
	assert.NoError(t, err)

	err = client.Save(&model2)
	assert.NoError(t, err)

	var matchingModels []*Model
	err = client.All(&matchingModels)
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 2)
	assertModelsEqual(t, &model1, matchingModels[0])
	assertModelsEqual(t, &model2, matchingModels[1])
}

func TestAllReturnsErrorIfArgumentInvalid(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var model Model
	err := client.All(&model)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intSlice []int
	err = client.All(&intSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intPtrSlice []*int
	err = client.All(&intPtrSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))
}

func TestFind(t *testing.T) {
	db := initDB(t)
	client := New(db)

	model := Model{
		Bool:   true,
		String: "abc",
	}

	err := client.Save(&model)
	assert.NoError(t, err)

	var matchingModel Model
	err = client.Find(
		&matchingModel,
		query.NewClause("where string = ?", "abc"))
	assert.NoError(t, err)

	assertModelsEqual(t, &model, &matchingModel)
}

func TestFindReturnsErrorIfArgumentInvalid(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var models []*Model
	err := client.Find(&models)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))

	var i int
	err = client.Find(&i)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))

	var intPtr *int
	err = client.Find(&intPtr)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))
}

func TestFindReturnsErrorIfModelNotFound(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var matchingModel Model
	err := client.Find(
		&matchingModel,
		query.NewClause("where string = ?", "abc"))
	assert.True(t, errors.Is(err, query.ErrModelNotFound))
}

func TestFindAll(t *testing.T) {
	db := initDB(t)
	client := New(db)

	model1 := Model{
		Bool:   true,
		String: "abc",
	}
	model2 := Model{
		Bool:   true,
		String: "def",
	}

	err := client.Save(&model1)
	assert.NoError(t, err)

	err = client.Save(&model2)
	assert.NoError(t, err)

	var matchingModels []*Model
	err = client.FindAll(
		&matchingModels,
		query.NewClause("where string = ?", "abc"))
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 1)
	assertModelsEqual(t, &model1, matchingModels[0])
}

func TestFindAllReturnsErrorIfArgumentInvalid(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var model Model
	err := client.FindAll(&model)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intSlice []int
	err = client.FindAll(&intSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intPtrSlice []*int
	err = client.FindAll(&intPtrSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))
}

func TestSaveReturnsErrorIfArgumentInvalid(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var models []*Model
	err := client.Save(&models)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))

	var i int
	err = client.Save(&i)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))

	var intPtr *int
	err = client.Save(&intPtr)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))
}

func TestSaveUpdatesExistingModel(t *testing.T) {
	db := initDB(t)
	client := New(db)

	model := Model{
		Bool:   true,
		String: "abc",
	}

	err := client.Save(&model)
	assert.NoError(t, err)

	model.String = "def"

	err = client.Save(&model)
	assert.NoError(t, err)

	var matchingModel Model
	err = client.Find(
		&matchingModel,
		query.NewClause("where string = ?", "def"))
	assert.NoError(t, err)

	assertModelsEqual(t, &model, &matchingModel)
}

func TestSaveSetsCreatedAtIfPresent(t *testing.T) {
	db := initDB(t)
	createManagedFieldsModelsTable(t, db)

	client := New(db)

	model := ManagedFieldsModel{
		Bool:   true,
		String: "abc",
	}

	err := client.Save(&model)
	assert.NoError(t, err)

	var matchingModel ManagedFieldsModel
	err = client.Find(
		&matchingModel,
		query.NewClause("where string = ?", "abc"))

	var zeroTime time.Time
	assert.NotEqual(t, matchingModel.CreatedAt, zeroTime)
}

func TestSaveUpdatesUpdatedAtIfPresent(t *testing.T) {
	db := initDB(t)
	createManagedFieldsModelsTable(t, db)

	client := New(db)

	model := ManagedFieldsModel{
		Bool:   true,
		String: "abc",
	}

	err := client.Save(&model)
	assert.NoError(t, err)

	var matchingModel ManagedFieldsModel
	err = client.Find(
		&matchingModel,
		query.NewClause("where string = ?", "abc"))

	var zeroTime time.Time
	assert.NotEqual(t, matchingModel.UpdatedAt, zeroTime)

	previousUpdatedAt := matchingModel.UpdatedAt

	err = client.Save(&model)
	assert.NoError(t, err)

	err = client.Find(
		&matchingModel,
		query.NewClause("where string = ?", "abc"))

	assert.True(t, matchingModel.UpdatedAt.After(previousUpdatedAt))
}
