package test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func InitDB(t *testing.T) *sql.DB {
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

	CreateModelsTable(t, db)

	return db
}

func CreateModelsTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS \"models\" (\"id\" integer primary key autoincrement,\"bool\" bool,\"string\" varchar(255));")
	assert.NoError(t, err)
}

func CreateManagedFieldsModelsTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS \"managed_fields_models\" (\"id\" integer primary key autoincrement,\"bool\" bool,\"string\" varchar(255),\"created_at\" datetime, \"updated_at\" datetime);")
	assert.NoError(t, err)
}

func AssertModelsEqual(t *testing.T, m1, m2 *Model) {
	assert.Equal(t, m1.Bool, m2.Bool)
	assert.Equal(t, m1.String, m2.String)
}
