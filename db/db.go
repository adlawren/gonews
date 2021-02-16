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

func scanFeed(rows *sql.Rows, f *feed.Feed) error {
	return rows.Scan(&f.ID, &f.URL, &f.FetchLimit)
}

func scanTag(rows *sql.Rows, t *feed.Tag) error {
	return rows.Scan(&t.ID, &t.Name, &t.FeedID)
}

func scanItem(rows *sql.Rows, i *feed.Item) error {
	return rows.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Title,
		&i.Description,
		&i.Link,
		&i.Published,
		&i.Hide,
		&i.FeedID)
}

func scanTimestamp(rows *sql.Rows, ts *timestamp.Timestamp) error {
	return rows.Scan(&ts.ID, &ts.T)
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

func (sdb *sqlDB) Feeds() ([]*feed.Feed, error) {
	var feeds []*feed.Feed

	rows, err := sdb.db.Query("select * from feeds;")
	defer rows.Close()
	if err != nil {
		return feeds, errors.Wrap(err, "failed to execute query")
	}

	for rows.Next() {
		var feed feed.Feed
		err = scanFeed(rows, &feed)
		if err != nil {
			return feeds, errors.Wrap(err, "failed to scan feed")
		}

		feeds = append(feeds, &feed)
	}

	err = rows.Err()
	if err != nil {
		return feeds, errors.Wrap(err, "cursor error")
	}

	return feeds, nil
}

func (sdb *sqlDB) FeedsFromTag(t *feed.Tag) ([]*feed.Feed, error) {
	var feeds []*feed.Feed

	stmt, err := sdb.db.Prepare("select * from feeds where id=(select feed_id from tags where name=?);")
	defer stmt.Close()
	if err != nil {
		return feeds, errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(t.Name)
	defer rows.Close()
	if err != nil {
		return feeds, errors.Wrap(err, "failed to execute prepared statement")
	}

	for rows.Next() {
		var feed feed.Feed
		err = scanFeed(rows, &feed)
		if err != nil {
			return feeds, errors.Wrap(err, "failed to scan feed")
		}

		feeds = append(feeds, &feed)
	}

	err = rows.Err()
	if err != nil {
		return feeds, errors.Wrap(err, "cursor error")
	}

	return feeds, nil
}

func (sdb *sqlDB) updateFeed(f *feed.Feed) error {
	stmt, err := sdb.db.Prepare("update feeds set url=?, fetch_limit=? where id=?;")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	res, err := stmt.Exec(f.URL, f.FetchLimit, f.ID)
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

func (sdb *sqlDB) insertFeed(f *feed.Feed) error {
	stmt, err := sdb.db.Prepare("insert into feeds (url, fetch_limit) values (?, ?);")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	res, err := stmt.Exec(f.URL, f.FetchLimit)
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

	f.ID = uint(id)

	return nil
}

func (sdb *sqlDB) MatchingFeed(f *feed.Feed) (*feed.Feed, error) {
	stmt, err := sdb.db.Prepare("select * from feeds where url=?;")
	defer stmt.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(f.URL)
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(
			err,
			"failed to execute prepared statement")
	}

	if !rows.Next() {
		return nil, nil
	}

	var feed feed.Feed
	err = scanFeed(rows, &feed)
	if err != nil {
		return &feed, errors.Wrap(err, "failed to scan feed")
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "cursor error")
	}

	return &feed, nil
}

func (sdb *sqlDB) SaveFeed(f *feed.Feed) error {
	stmt, err := sdb.db.Prepare("select count(*) from feeds where id=?;")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(f.ID)
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
		err = sdb.updateFeed(f)
	} else {
		err = sdb.insertFeed(f)
	}

	return errors.Wrap(err, "failed to save feed")
}

func (sdb *sqlDB) Tags() ([]*feed.Tag, error) {
	var tags []*feed.Tag

	rows, err := sdb.db.Query("select * from tags;")
	defer rows.Close()
	if err != nil {
		return tags, errors.Wrap(err, "failed to execute query")
	}

	for rows.Next() {
		var tag feed.Tag
		err = scanTag(rows, &tag)
		if err != nil {
			return tags, errors.Wrap(err, "failed to scan tag")
		}

		tags = append(tags, &tag)
	}

	err = rows.Err()
	if err != nil {
		return tags, errors.Wrap(err, "cursor error")
	}

	return tags, nil
}

func (sdb *sqlDB) MatchingTag(t *feed.Tag) (*feed.Tag, error) {
	stmt, err := sdb.db.Prepare("select * from tags where name=?;")
	defer stmt.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(t.Name)
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute prepared statement")
	}

	if !rows.Next() {
		return nil, nil
	}

	var tag feed.Tag
	err = scanTag(rows, &tag)
	if err != nil {
		return &tag, errors.Wrap(err, "failed to scan tag")
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "cursor error")
	}

	return &tag, nil
}

func (sdb *sqlDB) updateTag(t *feed.Tag) error {
	stmt, err := sdb.db.Prepare("update tags set name=?, feed_id=? where id=?;")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	res, err := stmt.Exec(t.Name, t.FeedID, t.ID)
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

func (sdb *sqlDB) insertTag(t *feed.Tag) error {
	stmt, err := sdb.db.Prepare("insert into tags (name, feed_id) values (?, ?);")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	res, err := stmt.Exec(t.Name, t.FeedID)
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

	t.ID = uint(id)

	return nil
}

func (sdb *sqlDB) SaveTag(t *feed.Tag) error {
	stmt, err := sdb.db.Prepare("select count(*) from tags where id=?;")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(t.ID)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "failed to execute prepared statement")
	}

	if !rows.Next() {
		return errors.New("cursor is empty")
	}

	var count int
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
		err = sdb.updateTag(t)
	} else {
		err = sdb.insertTag(t)
	}

	return errors.Wrap(err, "failed to save tag")
}

func (sdb *sqlDB) Items() ([]*feed.Item, error) {
	var items []*feed.Item
	rows, err := sdb.db.Query("select * from items order by published desc;")
	defer rows.Close()
	if err != nil {
		return items, errors.Wrap(err, "failed to execute query")
	}

	for rows.Next() {
		var item feed.Item
		err = scanItem(rows, &item)
		if err != nil {
			return items, errors.Wrap(err, "failed to scan item")
		}

		items = append(items, &item)
	}

	err = rows.Err()
	if err != nil {
		return items, errors.Wrap(err, "cursor error")
	}

	return items, nil
}

func (sdb *sqlDB) MatchingItem(i *feed.Item) (*feed.Item, error) {
	stmt, err := sdb.db.Prepare("select * from items where name=? and title=? and link=? limit 1;")
	defer stmt.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(i.Name, i.Title, i.Link)
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute prepared statement")
	}

	if !rows.Next() {
		return nil, nil
	}

	var item feed.Item
	err = scanItem(rows, &item)
	if err != nil {
		return &item, errors.Wrap(err, "failed to scan item")
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "cursor error")
	}

	return &item, nil
}

func (sdb *sqlDB) Item(id uint) (*feed.Item, error) {
	stmt, err := sdb.db.Prepare("select * from items where id=?;")
	defer stmt.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(id)
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(
			err,
			"failed to execute prepared statement")
	}

	if !rows.Next() {
		return nil, errors.New("cursor is empty")
	}

	var item feed.Item
	err = scanItem(rows, &item)
	if err != nil {
		return &item, errors.Wrap(err, "failed to scan item")
	}

	err = rows.Err()
	if err != nil {
		return &item, errors.Wrap(err, "cursor error")
	}

	return &item, nil
}

func (sdb *sqlDB) ItemsFromFeed(f *feed.Feed) ([]*feed.Item, error) {
	var items []*feed.Item

	stmt, err := sdb.db.Prepare("select * from items where feed_id=? order by published desc;")
	defer stmt.Close()
	if err != nil {
		return items, errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(f.ID)
	defer rows.Close()
	if err != nil {
		return items, errors.Wrap(
			err,
			"failed to execute prepared statement")
	}

	for rows.Next() {
		var item feed.Item
		err = scanItem(rows, &item)
		if err != nil {
			return items, errors.Wrap(err, "failed to scan item")
		}

		items = append(items, &item)
	}

	err = rows.Err()
	if err != nil {
		return items, errors.Wrap(err, "cursor error")
	}

	return items, nil
}

func (sdb *sqlDB) ItemsFromTag(t *feed.Tag) ([]*feed.Item, error) {
	var items []*feed.Item

	stmt, err := sdb.db.Prepare("select * from items where feed_id in (select feed_id from tags where name=?) order by published desc;")
	defer stmt.Close()
	if err != nil {
		return items, errors.Wrap(
			err,
			"failed to create prepared statement")
	}

	rows, err := stmt.Query(t.Name)
	defer rows.Close()
	if err != nil {
		return items, errors.Wrap(
			err,
			"failed to execute prepared statement")
	}

	for rows.Next() {
		var item feed.Item
		err = scanItem(rows, &item)
		if err != nil {
			return items, errors.Wrap(err, "failed to scan item")
		}

		items = append(items, &item)
	}

	err = rows.Err()
	if err != nil {
		return items, errors.Wrap(err, "cursor error")
	}

	return items, nil
}

func (sdb *sqlDB) updateItem(i *feed.Item) error {
	stmt, err := sdb.db.Prepare("update items set name=?, email=?, title=?, description=?, link=?, published=?, hide=?, feed_id=? where id=?;")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	res, err := stmt.Exec(i.Name, i.Email, i.Title, i.Description, i.Link, i.Published, i.Hide, i.FeedID, i.ID)
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

func (sdb *sqlDB) insertItem(i *feed.Item) error {
	stmt, err := sdb.db.Prepare("insert into items (name, email, title, description, link, published, hide, feed_id) values (?, ?, ?, ?, ?, ?, ?, ?);")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	res, err := stmt.Exec(i.Name, i.Email, i.Title, i.Description, i.Link, i.Published, i.Hide, i.FeedID)
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

	i.ID = uint(id)

	return nil
}

func (sdb *sqlDB) SaveItem(i *feed.Item) error {
	stmt, err := sdb.db.Prepare("select count(*) from items where id=?;")
	defer stmt.Close()
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	rows, err := stmt.Query(i.ID)
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
		err = sdb.updateItem(i)
	} else {
		err = sdb.insertItem(i)
	}

	return errors.Wrap(err, "failed to save item")
}

func (sdb *sqlDB) Close() error {
	err := sdb.db.Close()
	return errors.Wrap(err, "failed to close DB")
}
