package client

import (
	"database/sql"
	"errors"
	"fmt"
	"gonews/db/orm/query"
	"gonews/db/orm/query/clause"
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

type IdMissingModel struct {
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

func TestAllReturnsErrorIfIdIsMissing(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var models []*IdMissingModel
	err := client.All(&models)
	assert.True(t, errors.Is(err, query.ErrMissingIdField))
}

func TestDeleteAll(t *testing.T) {
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
	model3 := Model{
		Bool:   true,
		String: "ghi",
	}

	err := client.Save(&model1)
	assert.NoError(t, err)

	err = client.Save(&model2)
	assert.NoError(t, err)

	err = client.Save(&model3)
	assert.NoError(t, err)

	err = client.DeleteAll(&[]*Model{&model1, &model2})
	assert.NoError(t, err)

	var models []*Model
	err = client.All(&models)
	assert.Len(t, models, 1)
	assertModelsEqual(t, &model3, models[0])
}

func TestDeleteAllReturnsErrorIfArgumentInvalid(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var model Model
	err := client.DeleteAll(&model)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intSlice []int
	err = client.DeleteAll(&intSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intPtrSlice []*int
	err = client.DeleteAll(&intPtrSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))
}

func TestDeleteAllReturnsErrorIfIdIsMissing(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var models []*IdMissingModel
	err := client.DeleteAll(&models)
	assert.True(t, errors.Is(err, query.ErrMissingIdField))
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
		clause.New("where string = ?", "abc"))
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

func TestFindReturnsErrorIfIdIsMissing(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var model IdMissingModel
	err := client.Find(&model)
	assert.True(t, errors.Is(err, query.ErrMissingIdField))
}

func TestFindReturnsErrorIfModelNotFound(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var matchingModel Model
	err := client.Find(
		&matchingModel,
		clause.New("where string = ?", "abc"))
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
		clause.New("where string = ?", "abc"))
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

func TestFindAllReturnsErrorIfIdIsMissing(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var models []*IdMissingModel
	err := client.FindAll(&models)
	assert.True(t, errors.Is(err, query.ErrMissingIdField))
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

func TestSaveReturnsErrorIfIdIsMissing(t *testing.T) {
	db := initDB(t)
	client := New(db)

	var model IdMissingModel
	err := client.Save(&model)
	assert.True(t, errors.Is(err, query.ErrMissingIdField))
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
		clause.New("where string = ?", "def"))
	assert.NoError(t, err)

	assertModelsEqual(t, &model, &matchingModel)
}

func TestSaveSetsID(t *testing.T) {
	db := initDB(t)
	client := New(db)

	model := Model{
		Bool:   true,
		String: "abc",
	}

	err := client.Save(&model)
	assert.NoError(t, err)

	assert.NotZero(t, model.ID)
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

	var zeroTime time.Time
	assert.NotEqual(t, model.CreatedAt, zeroTime)
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

	var zeroTime time.Time
	assert.NotEqual(t, model.UpdatedAt, zeroTime)

	previousUpdatedAt := model.UpdatedAt

	err = client.Save(&model)
	assert.NoError(t, err)

	assert.True(t, model.UpdatedAt.After(previousUpdatedAt))
}
