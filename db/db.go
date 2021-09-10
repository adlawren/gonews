package db

import (
	"database/sql"
	"fmt"
	"gonews/config"
	"gonews/db/orm/client"
	"gonews/db/orm/query"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose"
)

// DB contains the methods needed to store and read data from the underlying
// database
type DB interface {
	Ping() error
	Migrate(string) error
	All(interface{}) error
	Find(interface{}, ...*query.Clause) error
	FindAll(interface{}, ...*query.Clause) error
	Save(interface{}) error
	Close() error
}

// New creates a struct which supports the operations in the DB interface
func New(cfg *config.DBConfig) (DB, error) {
	db, err := sql.Open("sqlite3", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	return &sqlDB{db: db}, nil
}

type sqlDB struct {
	db *sql.DB
}

func (sdb *sqlDB) Ping() error {
	err := sdb.db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping DB: %w", err)
	}

	return nil
}

func (sdb *sqlDB) Migrate(migrationsDir string) error {
	_, err := os.Stat(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to stat migrations directory: %w", err)
	}

	err = goose.SetDialect("sqlite3")
	if err != nil {
		return fmt.Errorf("failed to set goose DB driver: %w", err)
	}

	err = goose.Up(sdb.db, migrationsDir)
	if err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}

	return nil
}

func (sdb *sqlDB) client() client.Client {
	return client.New(sdb.db)
}

func (sdb *sqlDB) All(ptr interface{}) error {
	return sdb.client().All(ptr)
}

func (sdb *sqlDB) Find(ptr interface{}, clauses ...*query.Clause) error {
	return sdb.client().Find(ptr, clauses...)
}

func (sdb *sqlDB) FindAll(ptr interface{}, clauses ...*query.Clause) error {
	return sdb.client().FindAll(ptr, clauses...)
}

func (sdb *sqlDB) Save(ptr interface{}) error {
	return sdb.client().Save(ptr)
}

func (sdb *sqlDB) Close() error {
	err := sdb.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close DB: %w", err)
	}

	return nil
}
