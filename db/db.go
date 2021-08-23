package db

import (
	"database/sql"
	"gonews/config"
	"gonews/db/orm"
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
	MatchingTimestamp(*timestamp.Timestamp) (*timestamp.Timestamp, error)
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

func (sdb *sqlDB) all(ptr interface{}) error {
	return sdb.findAll(ptr)
}

func (sdb *sqlDB) find(ptr interface{}, clauses ...*orm.QueryClause) error {
	query, err := orm.SelectModel(ptr, clauses...)
	if err != nil {
		return errors.Wrap(err, "failed to create query")
	}

	return errors.Wrap(query.Exec(sdb.db), "failed to execute query")
}

func (sdb *sqlDB) findAll(ptr interface{}, clauses ...*orm.QueryClause) error {
	query, err := orm.SelectModels(ptr, clauses...)
	if err != nil {
		return errors.Wrap(err, "failed to create query")
	}

	return errors.Wrap(query.Exec(sdb.db), "failed to execute query")
}

func (sdb *sqlDB) save(ptr interface{}) error {
	query, err := orm.UpsertModel(ptr)
	if err != nil {
		return errors.Wrap(err, "failed to create query")
	}

	return errors.Wrap(query.Exec(sdb.db), "failed to execute query")
}

func (sdb *sqlDB) MatchingTimestamp(ts *timestamp.Timestamp) (*timestamp.Timestamp, error) {
	var timestamp timestamp.Timestamp
	err := sdb.find(&timestamp, orm.Clause("where name = ?", ts.Name))
	if errors.Is(err, orm.ErrModelNotFound) {
		return nil, nil
	}

	return &timestamp, errors.Wrap(err, "failed to get matching timestamp")
}

func (sdb *sqlDB) SaveTimestamp(ts *timestamp.Timestamp) error {
	return errors.Wrap(sdb.save(ts), "failed to save timestamp")
}

func (sdb *sqlDB) Users() ([]*user.User, error) {
	var users []*user.User
	return users, errors.Wrap(sdb.all(&users), "failed to get users")
}

func (sdb *sqlDB) MatchingUser(u *user.User) (*user.User, error) {
	var user user.User
	err := sdb.find(&user, orm.Clause("where username = ?", u.Username))
	if errors.Is(err, orm.ErrModelNotFound) {
		return nil, nil
	}

	return &user, errors.Wrap(err, "failed to get matching user")
}

func (sdb *sqlDB) SaveUser(u *user.User) error {
	return errors.Wrap(sdb.save(u), "failed to save user")
}

func (sdb *sqlDB) Feeds() ([]*feed.Feed, error) {
	var feeds []*feed.Feed
	return feeds, errors.Wrap(sdb.all(&feeds), "failed to get feeds")
}

func (sdb *sqlDB) FeedsFromTag(t *feed.Tag) ([]*feed.Feed, error) {
	var feeds []*feed.Feed
	err := sdb.findAll(&feeds, orm.Clause("where tag_id in (select id from tags where name = ?)", t.Name))
	return feeds, errors.Wrap(err, "failed to find feeds")
}

func (sdb *sqlDB) MatchingFeed(f *feed.Feed) (*feed.Feed, error) {
	var feed feed.Feed
	err := sdb.find(&feed, orm.Clause("where url = ?", f.URL))
	if errors.Is(err, orm.ErrModelNotFound) {
		return nil, nil
	}

	return &feed, errors.Wrap(err, "failed to get matching feed")
}

func (sdb *sqlDB) SaveFeed(f *feed.Feed) error {
	return errors.Wrap(sdb.save(f), "failed to save feed")
}

func (sdb *sqlDB) Tags() ([]*feed.Tag, error) {
	var tags []*feed.Tag
	return tags, errors.Wrap(sdb.all(&tags), "failed to get all tags")
}

func (sdb *sqlDB) MatchingTag(t *feed.Tag) (*feed.Tag, error) {
	var tag feed.Tag
	err := sdb.find(&tag, orm.Clause("where name = ?", t.Name))
	if errors.Is(err, orm.ErrModelNotFound) {
		return nil, nil
	}

	return &tag, errors.Wrap(err, "failed to get matching tag")
}

func (sdb *sqlDB) SaveTag(t *feed.Tag) error {
	return errors.Wrap(sdb.save(t), "failed to save tag")
}

func (sdb *sqlDB) Items() ([]*feed.Item, error) {
	var items []*feed.Item
	return items, errors.Wrap(sdb.all(&items), "failed to get all items")
}

func (sdb *sqlDB) MatchingItem(i *feed.Item) (*feed.Item, error) {
	var item feed.Item
	err := sdb.find(&item, orm.Clause("where link = ?", i.Link))
	if errors.Is(err, orm.ErrModelNotFound) {
		return nil, nil
	}

	return &item, errors.Wrap(err, "failed to get matching item")
}

func (sdb *sqlDB) Item(id uint) (*feed.Item, error) {
	var item feed.Item
	err := sdb.find(&item, orm.Clause("where id = ?", id))
	return &item, errors.Wrap(err, "failed to find item")
}

func (sdb *sqlDB) ItemsFromFeed(f *feed.Feed) ([]*feed.Item, error) {
	var items []*feed.Item
	err := sdb.findAll(&items, orm.Clause("where feed_id = ?", f.ID))
	return items, errors.Wrap(err, "failed to find items")
}

func (sdb *sqlDB) ItemsFromTag(t *feed.Tag) ([]*feed.Item, error) {
	var items []*feed.Item
	err := sdb.findAll(&items, orm.Clause("where feed_id in (select feed_id from tags where name = ?)", t.Name))
	return items, errors.Wrap(err, "failed to find items")
}

func (sdb *sqlDB) SaveItem(i *feed.Item) error {
	return errors.Wrap(sdb.save(i), "failed to save item")
}

func (sdb *sqlDB) Close() error {
	err := sdb.db.Close()
	return errors.Wrap(err, "failed to close DB")
}
