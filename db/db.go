package db

import (
	"database/sql"
	"gonews/config"
	"gonews/feed"
	"gonews/timestamp"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/pressly/goose"
)

// DB contains the methods needed to store and read data from the underlying
// database
type DB interface {
	Ping() error
	Migrate(string) error
	Timestamp() (*time.Time, error)
	SaveTimestamp(*time.Time) error
	Feeds() ([]*feed.Feed, error)
	FeedsFromTag(*feed.Tag) ([]*feed.Feed, error)
	MatchingFeed(*feed.Feed) (*feed.Feed, error)
	SaveFeed(*feed.Feed) error
	Tags() ([]*feed.Tag, error)
	MatchingTag(*feed.Tag) (*feed.Tag, error)
	SaveTag(*feed.Tag) error
	Items() ([]*feed.Item, error)
	ItemsFromFeed(*feed.Feed) ([]*feed.Item, error)
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

	return errors.Wrap(goose.Up(sdb.db, migrationsDir), "migrations failed")
}

func (sdb *sqlDB) Timestamp() (*time.Time, error) {
	rows, err := sdb.db.Query("select id, t from timestamps;")
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	var ts timestamp.Timestamp
	if rows.Next() {
		err = rows.Scan(&ts.ID, &ts.T)
	}

	return &ts.T, errors.Wrap(err, "failed to scan timestamp")
}

func (sdb *sqlDB) insertTimestamp(ts *timestamp.Timestamp) error {
	res, err := sdb.db.Exec("insert into timestamps (t) values (?);", ts.T)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		return errors.Wrap(err, "more than one row affected")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "failed to get id")
	}

	ts.ID = uint(id)

	return nil
}

func (sdb *sqlDB) updateTimestamp(ts *timestamp.Timestamp) error {
	res, err := sdb.db.Exec("update timestamps set t=? where id=?;", ts.T, ts.ID)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		return errors.Wrap(err, "more than one row affected")
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
		return errors.New("no rows")
	}
	err = rows.Scan(&count)
	if err != nil {
		return errors.Wrap(err, "scan failed")
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
		err = rows.Scan(&ts.ID, &ts.T)
	}
	if err != nil {
		return errors.Wrap(err, "scan failed")
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

func (sdb *sqlDB) Feeds() ([]*feed.Feed, error) {
	var feeds []*feed.Feed

	rows, err := sdb.db.Query("select id, url from feeds;")
	defer rows.Close()
	if err != nil {
		return feeds, errors.Wrap(err, "failed to get feeds")
	}

	for rows.Next() {
		if rows.Err() != nil {
			return feeds, errors.Wrap(rows.Err(), "failed to get feeds")
		}

		var feed feed.Feed
		err = rows.Scan(&feed.ID, &feed.URL)
		if err != nil {
			return feeds, err
		}

		feeds = append(feeds, &feed)
	}

	return feeds, nil
}

func (sdb *sqlDB) FeedsFromTag(t *feed.Tag) ([]*feed.Feed, error) {
	var feeds []*feed.Feed

	stmt, err := sdb.db.Prepare("select id, url from feeds where id=(select feed_id from tags where name=?);")
	defer stmt.Close()
	if err != nil {
		return feeds, errors.Wrap(err, "failed to get feeds")
	}

	rows, err := stmt.Query(t.Name)
	defer rows.Close()
	if err != nil {
		return feeds, errors.Wrap(err, "failed to get feeds")
	}

	for rows.Next() {
		if rows.Err() != nil {
			return feeds, errors.Wrap(rows.Err(), "failed to get feeds")
		}

		var feed feed.Feed
		err = rows.Scan(&feed.ID, &feed.URL)
		if err != nil {
			return feeds, err
		}

		feeds = append(feeds, &feed)
	}

	return feeds, nil
}

func (sdb *sqlDB) updateFeed(f *feed.Feed) error {
	res, err := sdb.db.Exec("update feeds set url=? where id=?;", f.URL, f.ID)
	if err != nil {
		return errors.Wrap(err, "failed to update feed")
	}
	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		return errors.Wrap(err, "failed to insert feed")
	}

	return nil
}

func (sdb *sqlDB) insertFeed(f *feed.Feed) error {
	res, err := sdb.db.Exec("insert into feeds (url) values (?);", f.URL)
	if err != nil {
		return errors.Wrap(err, "failed to insert feed")
	}
	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		return errors.Wrap(err, "failed to insert feed")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "failed to get id")
	}

	f.ID = uint(id)

	return nil
}

func (sdb *sqlDB) MatchingFeed(f *feed.Feed) (*feed.Feed, error) {
	stmt, err := sdb.db.Prepare("select id, url from feeds where url=?;")
	defer stmt.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(f.URL)
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	if !rows.Next() {
		return nil, nil
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "cursor returned error")
	}

	var feed feed.Feed
	err = rows.Scan(&feed.ID, &feed.URL)
	if err != nil {
		return &feed, errors.Wrap(err, "scan failed")
	}

	return &feed, nil
}

func (sdb *sqlDB) SaveFeed(f *feed.Feed) error {
	rows, err := sdb.db.Query("select count(*) from feeds where id=?;", f.ID)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	var count int
	if !rows.Next() {
		return errors.New("failed to get feed count")
	}
	err = rows.Scan(&count)
	if err != nil {
		return errors.Wrap(err, "scan failed")
	}

	err = rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close cursor")
	}

	if count != 0 {
		err = sdb.updateFeed(f)
	} else {
		err = sdb.insertFeed(f)
	}

	return errors.Wrap(err, "failed to save feed")
}

func (sdb *sqlDB) Tags() ([]*feed.Tag, error) {
	var tags []*feed.Tag

	rows, err := sdb.db.Query("select id, name, feed_id from tags;")
	defer rows.Close()
	if err != nil {
		return tags, errors.Wrap(err, "failed to get tags")
	}

	for rows.Next() {
		if rows.Err() != nil {
			return tags, errors.Wrap(rows.Err(), "failed to get tags")
		}

		var tag feed.Tag
		err = rows.Scan(&tag.ID, &tag.Name, &tag.FeedID)
		if err != nil {
			return tags, err
		}

		tags = append(tags, &tag)
	}

	return tags, nil
}

func (sdb *sqlDB) MatchingTag(t *feed.Tag) (*feed.Tag, error) {
	stmt, err := sdb.db.Prepare("select id, name from tags where name=?;")
	defer stmt.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(t.Name)
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	if !rows.Next() {
		return nil, nil
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "cursor returned error")
	}

	var tag feed.Tag
	err = rows.Scan(&tag.ID, &tag.Name)
	if err != nil {
		return &tag, errors.Wrap(err, "scan failed")
	}

	return &tag, nil
}

func (sdb *sqlDB) updateTag(t *feed.Tag) error {
	res, err := sdb.db.Exec("update tags set name=?, feed_id=? where id=?;", t.Name, t.FeedID, t.ID)
	if err != nil {
		return errors.Wrap(err, "failed to update tag")
	}

	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		return errors.Wrap(err, "failed to update tag")
	}

	return nil
}

func (sdb *sqlDB) insertTag(t *feed.Tag) error {
	res, err := sdb.db.Exec("insert into tags (name, feed_id) values (?, ?);", t.Name, t.FeedID)
	if err != nil {
		return errors.Wrap(err, "failed to insert tag")
	}

	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		return errors.Wrap(err, "failed to insert tag")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "failed to get id")
	}

	t.ID = uint(id)

	return nil
}

func (sdb *sqlDB) SaveTag(t *feed.Tag) error {
	rows, err := sdb.db.Query("select count(*) from tags where id=?;", t.ID)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "query failed")
	}

	if !rows.Next() {
		return errors.New("failed to get tag count")
	}

	var count int
	err = rows.Scan(&count)
	if err != nil {
		return errors.Wrap(err, "scan failed")
	}

	err = rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close cursor")
	}

	if count != 0 {
		err = sdb.updateTag(t)
	} else {
		err = sdb.insertTag(t)
	}

	return errors.Wrap(err, "failed to save tag")
}

func (sdb *sqlDB) Items() ([]*feed.Item, error) {
	var items []*feed.Item
	rows, err := sdb.db.Query("select id, name, email, title, description, link, published, feed_id from items order by published;")
	defer rows.Close()
	if err != nil {
		return items, errors.Wrap(err, "failed to get items")
	}

	for rows.Next() {
		if rows.Err() != nil {
			return items, errors.Wrap(rows.Err(), "failed to get items")
		}

		var item feed.Item
		err = rows.Scan(&item.ID, &item.Name, &item.Email, &item.Title, &item.Description, &item.Link, &item.Published, &item.FeedID)
		if err != nil {
			return items, err
		}

		items = append(items, &item)
	}

	return items, nil
}

func (sdb *sqlDB) MatchingItem(i *feed.Item) (*feed.Item, error) {
	stmt, err := sdb.db.Prepare("select id, name, email, title, description, link, published, feed_id from items where name=? and title=? and link=? limit 1;")
	defer stmt.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(i.Name, i.Title, i.Link)
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	if !rows.Next() {
		return nil, nil
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "cursor returned error")
	}

	var item feed.Item
	err = rows.Scan(
		&item.ID,
		&item.Name,
		&item.Email,
		&item.Title,
		&item.Description,
		&item.Link,
		&item.Published,
		&item.FeedID)
	if err != nil {
		return &item, errors.Wrap(err, "scan failed")
	}

	return &item, nil
}

func (sdb *sqlDB) ItemsFromFeed(f *feed.Feed) ([]*feed.Item, error) {
	var items []*feed.Item

	stmt, err := sdb.db.Prepare("select id, name, email, title, description, link, published, feed_id from items where feed_id=? order by published;")
	defer stmt.Close()
	if err != nil {
		return items, errors.Wrap(err, "failed to get items")
	}

	rows, err := stmt.Query(f.ID)
	defer rows.Close()
	if err != nil {
		return items, errors.Wrap(err, "failed to get items")
	}

	for rows.Next() {
		if rows.Err() != nil {
			return items, errors.Wrap(rows.Err(), "failed to get items")
		}

		var item feed.Item
		err = rows.Scan(&item.ID, &item.Name, &item.Email, &item.Title, &item.Description, &item.Link, &item.Published, &item.FeedID)
		if err != nil {
			return items, err
		}

		items = append(items, &item)
	}

	return items, nil
}

func (sdb *sqlDB) updateItem(i *feed.Item) error {
	res, err := sdb.db.Exec("update items set name=?, email=?, title=?, description=?, link=?, published=?, hide=?, feed_id=? where id=?;", i.Name, i.Email, i.Title, i.Description, i.Link, i.Published, i.Hide, i.FeedID, i.ID)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		return errors.Wrap(err, "more than one row affected")
	}

	return nil
}

func (sdb *sqlDB) insertItem(i *feed.Item) error {
	res, err := sdb.db.Exec("insert into items (name, email, title, description, link, published, hide, feed_id) values (?, ?, ?, ?, ?, ?, ?, ?);", i.Name, i.Email, i.Title, i.Description, i.Link, i.Published, i.Hide, i.FeedID, i.ID)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		return errors.Wrap(err, "more than one row affected")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "failed to get id")
	}

	i.ID = uint(id)

	return nil
}

func (sdb *sqlDB) SaveItem(i *feed.Item) error {
	rows, err := sdb.db.Query("select count(*) from items where id=?;", i.ID)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	var count int
	if !rows.Next() {
		return errors.New("failed to get item count")
	}
	err = rows.Scan(&count)
	if err != nil {
		return errors.Wrap(err, "scan failed")
	}

	err = rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close cursor")
	}

	if count != 0 {
		err = sdb.updateItem(i)
	} else {
		err = sdb.insertItem(i)
	}

	return errors.Wrap(err, "failed to save item")
}

func (sdb *sqlDB) Close() error {
	err := sdb.db.Close()
	return errors.Wrap(err, "unable to close db")
}
