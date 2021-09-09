package db

import (
	"database/sql"
	"gonews/config"
	"gonews/db/orm/client"
	"gonews/db/orm/query"
	"gonews/feed"
	"gonews/timestamp"
	"gonews/user"
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
	Find(interface{}, ...*query.Clause) error
	Save(interface{}) error
	SaveTimestamp(*timestamp.Timestamp) error
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

func (sdb *sqlDB) client() client.Client {
	return client.New(sdb.db)
}

func (sdb *sqlDB) Find(ptr interface{}, clauses ...*query.Clause) error {
	return sdb.client().Find(ptr, clauses...)
}

func (sdb *sqlDB) Save(ptr interface{}) error {
	return sdb.client().Save(ptr)
}

func (sdb *sqlDB) SaveTimestamp(ts *timestamp.Timestamp) error {
	return errors.Wrap(sdb.client().Save(ts), "failed to save timestamp")
}

func (sdb *sqlDB) Users() ([]*user.User, error) {
	var users []*user.User
	return users, errors.Wrap(sdb.client().All(&users), "failed to get users")
}

func (sdb *sqlDB) MatchingUser(u *user.User) (*user.User, error) {
	var user user.User
	err := sdb.client().Find(&user, query.NewClause("where username = ?", u.Username))
	if errors.Is(err, query.ErrModelNotFound) {
		return nil, nil
	}

	return &user, errors.Wrap(err, "failed to get matching user")
}

func (sdb *sqlDB) SaveUser(u *user.User) error {
	return errors.Wrap(sdb.client().Save(u), "failed to save user")
}

func (sdb *sqlDB) Feeds() ([]*feed.Feed, error) {
	var feeds []*feed.Feed
	return feeds, errors.Wrap(sdb.client().All(&feeds), "failed to get feeds")
}

func (sdb *sqlDB) FeedsFromTag(t *feed.Tag) ([]*feed.Feed, error) {
	var feeds []*feed.Feed
	err := sdb.client().FindAll(&feeds, query.NewClause("where tag_id in (select id from tags where name = ?)", t.Name))
	return feeds, errors.Wrap(err, "failed to find feeds")
}

func (sdb *sqlDB) MatchingFeed(f *feed.Feed) (*feed.Feed, error) {
	var feed feed.Feed
	err := sdb.client().Find(&feed, query.NewClause("where url = ?", f.URL))
	if errors.Is(err, query.ErrModelNotFound) {
		return nil, nil
	}

	return &feed, errors.Wrap(err, "failed to get matching feed")
}

func (sdb *sqlDB) SaveFeed(f *feed.Feed) error {
	return errors.Wrap(sdb.client().Save(f), "failed to save feed")
}

func (sdb *sqlDB) Tags() ([]*feed.Tag, error) {
	var tags []*feed.Tag
	return tags, errors.Wrap(sdb.client().All(&tags), "failed to get all tags")
}

func (sdb *sqlDB) MatchingTag(t *feed.Tag) (*feed.Tag, error) {
	var tag feed.Tag
	err := sdb.client().Find(&tag, query.NewClause("where name = ?", t.Name))
	if errors.Is(err, query.ErrModelNotFound) {
		return nil, nil
	}

	return &tag, errors.Wrap(err, "failed to get matching tag")
}

func (sdb *sqlDB) SaveTag(t *feed.Tag) error {
	return errors.Wrap(sdb.client().Save(t), "failed to save tag")
}

func (sdb *sqlDB) Items() ([]*feed.Item, error) {
	var items []*feed.Item
	return items, errors.Wrap(sdb.client().All(&items), "failed to get all items")
}

func (sdb *sqlDB) MatchingItem(i *feed.Item) (*feed.Item, error) {
	var item feed.Item
	err := sdb.client().Find(&item, query.NewClause("where link = ?", i.Link))
	if errors.Is(err, query.ErrModelNotFound) {
		return nil, nil
	}

	return &item, errors.Wrap(err, "failed to get matching item")
}

func (sdb *sqlDB) Item(id uint) (*feed.Item, error) {
	var item feed.Item
	err := sdb.client().Find(&item, query.NewClause("where id = ?", id))
	return &item, errors.Wrap(err, "failed to find item")
}

func (sdb *sqlDB) ItemsFromFeed(f *feed.Feed) ([]*feed.Item, error) {
	var items []*feed.Item
	err := sdb.client().FindAll(&items, query.NewClause("where feed_id = ?", f.ID))
	return items, errors.Wrap(err, "failed to find items")
}

func (sdb *sqlDB) ItemsFromTag(t *feed.Tag) ([]*feed.Item, error) {
	var items []*feed.Item
	err := sdb.client().FindAll(&items, query.NewClause("where feed_id in (select feed_id from tags where name = ?)", t.Name))
	return items, errors.Wrap(err, "failed to find items")
}

func (sdb *sqlDB) SaveItem(i *feed.Item) error {
	return errors.Wrap(sdb.client().Save(i), "failed to save item")
}

func (sdb *sqlDB) Close() error {
	err := sdb.db.Close()
	return errors.Wrap(err, "failed to close DB")
}
