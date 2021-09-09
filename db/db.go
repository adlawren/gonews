package db

import (
	"database/sql"
	"gonews/config"
	"gonews/db/orm/client"
	"gonews/db/orm/query"
	"gonews/feed"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
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

func (sdb *sqlDB) MatchingItem(i *feed.Item) (*feed.Item, error) {
	var item feed.Item
	err := sdb.client().Find(&item, query.NewClause("where link = ?", i.Link))
	if errors.Is(err, query.ErrModelNotFound) {
		return nil, nil
	}

	return &item, errors.Wrap(err, "failed to get matching item")
}

func (sdb *sqlDB) SaveItem(i *feed.Item) error {
	return errors.Wrap(sdb.client().Save(i), "failed to save item")
}

func (sdb *sqlDB) Close() error {
	err := sdb.db.Close()
	return errors.Wrap(err, "failed to close DB")
}
