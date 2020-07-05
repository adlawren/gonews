package db

import (
	"database/sql"
	"gonews/config"
	"gonews/feed"
	"gonews/timestamp"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	//. "github.com/volatiletech/sqlboiler/queries/qm"
)

// DB contains the methods needed to store and read data from the underlying
// database
type DB interface {
	Ping() error
	Timestamp() (*time.Time, error)
	UpdateTimestamp(*time.Time) error
	AllFeeds() ([]*feed.Feed, error) // TODO: rm?
	//MatchingFeed(*feed.Feed) (*feed.Feed, error) // TODO: rm?
	FeedsFromTag(*feed.Tag) ([]*feed.Feed, error)
	SaveFeed(*feed.Feed) error
	AllTags() ([]*feed.Tag, error)
	MatchingTag(*feed.Tag) (*feed.Tag, error)
	SaveTagToFeed(*feed.Tag, *feed.Feed) error
	AllItems() ([]*feed.Item, error)
	MatchingItem(*feed.Item) (*feed.Item, error)
	ItemsFromFeed(*feed.Feed) ([]*feed.Item, error)
	SaveItemToFeed(*feed.Item, *feed.Feed) error
	//SaveItem(*feed.Item) error // TODO: rm?
	UpdateItem(*feed.Item) error
	Close() error
}

// New creates a struct which supports the operations in the DB interface
func New(cfg *config.DBConfig) (DB, error) {
	// db, err := gorm.Open("sqlite3", cfg.Path)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to open db")
	// }

	// return &gormDB{db: db}, nil

	db, err := sql.Open("sqlite3", cfg.DSN)
	return &sqlDB{db: db}, errors.Wrap(err, "failed to open DB")
}

type gormDB struct {
	db *gorm.DB
}

// TODO?
func (gdb *gormDB) Ping() error {
	return nil
}

func (gdb *gormDB) Timestamp() (*time.Time, error) {
	var t time.Time

	var timestamps []*timestamp.Timestamp
	err := gdb.db.Find(&timestamps, &timestamp.Timestamp{}).Error
	if err != nil {
		return &t, errors.Wrap(err, "failed to get timestamp")
	}

	if len(timestamps) == 1 {
		t = timestamps[0].T
	} else if len(timestamps) > 1 {
		return &t, errors.New("more than one timestamp in db")
	}

	return &t, nil
}

func (gdb *gormDB) UpdateTimestamp(t *time.Time) error {
	var ts timestamp.Timestamp
	err := gdb.db.FirstOrCreate(&ts, &timestamp.Timestamp{}).Error

	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	ts.T = *t

	err = gdb.db.Save(&ts).Error
	return errors.Wrap(err, "failed to save timestamp")
}

func (gdb *gormDB) AllFeeds() ([]*feed.Feed, error) {
	var feeds []*feed.Feed
	err := gdb.db.Find(&feeds, &feed.Feed{}).Error
	return feeds, errors.Wrap(err, "failed to get feeds")
}

func (gdb *gormDB) MatchingFeed(f *feed.Feed) (*feed.Feed, error) {
	var feed feed.Feed
	err := gdb.db.First(&feed, f).Error
	return &feed, errors.Wrap(err, "failed to get feed")
}

func (gdb *gormDB) FeedsFromTag(t *feed.Tag) ([]*feed.Feed, error) {
	var feeds []*feed.Feed
	err := gdb.db.Model(t).Related(&feeds).Error
	return feeds, errors.Wrap(err, "failed to get feeds")
}

func (gdb *gormDB) SaveFeed(f *feed.Feed) error {
	var existingFeed feed.Feed
	err := gdb.db.FirstOrCreate(&existingFeed, f).Error
	return errors.Wrap(err, "failed to save feed")
}

// TODO?
func (gdb *gormDB) AllTags() ([]*feed.Tag, error) {
	return nil, nil
}

func (gdb *gormDB) MatchingTag(t *feed.Tag) (*feed.Tag, error) {
	var tag feed.Tag
	err := gdb.db.First(&tag, t).Error
	return &tag, errors.Wrap(err, "failed to get tag")
}

func (gdb *gormDB) SaveTagToFeed(t *feed.Tag, f *feed.Feed) error {
	var existingFeed feed.Feed
	err := gdb.db.Find(&existingFeed, f).Error
	if err != nil {
		return errors.Wrap(err, "failed to get feed")
	}

	t.FeedID = existingFeed.ID

	var existingTag feed.Tag
	err = gdb.db.FirstOrCreate(&existingTag, t).Error
	return errors.Wrap(err, "failed to save tag")
}

func (gdb *gormDB) AllItems() ([]*feed.Item, error) {
	var items []*feed.Item
	err := gdb.db.Find(&items, &feed.Item{}).Error
	return items, errors.Wrap(err, "failed to get items")
}

func (gdb *gormDB) MatchingItem(i *feed.Item) (*feed.Item, error) {
	var items []*feed.Item
	err := gdb.db.Find(&items, i).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to get matching item")
	}
	if len(items) > 0 {
		return items[0], nil
	}

	return &feed.Item{}, nil
}

func (gdb *gormDB) ItemsFromFeed(f *feed.Feed) ([]*feed.Item, error) {
	var items []*feed.Item
	err := gdb.db.Model(f).Related(&items).Error
	return items, errors.Wrap(err, "failed to get items")
}

func (gdb *gormDB) SaveItemToFeed(i *feed.Item, f *feed.Feed) error {
	var existingFeed feed.Feed
	err := gdb.db.Find(&existingFeed, f).Error
	if err != nil {
		return errors.Wrap(err, "failed to get feed")
	}

	i.FeedID = existingFeed.ID

	return gdb.SaveItem(i)
}

func (gdb *gormDB) SaveItem(i *feed.Item) error {
	err := gdb.db.Save(i).Error
	return errors.Wrap(err, "failed to save item")
}

func (gdb *gormDB) UpdateItem(i *feed.Item) error {
	err := gdb.db.Model(&feed.Item{}).Update(i).Error
	return errors.Wrap(err, "failed to update item")
}

func (gdb *gormDB) Close() error {
	err := gdb.db.Close()
	return errors.Wrap(err, "unable to close db")
}

// TODO
type sqlDB struct {
	db *sql.DB
}

func (sdb *sqlDB) Ping() error {
	return errors.Wrap(sdb.db.Ping(), "failed to ping DB")
}

func (sdb *sqlDB) Timestamp() (*time.Time, error) {
	// var t time.Time

	// var timestamps []*timestamp.Timestamp
	// err := sdb.db.Find(&timestamps, &timestamp.Timestamp{}).Error
	// if err != nil {
	// 	return &t, errors.Wrap(err, "failed to get timestamp")
	// }

	// if len(timestamps) == 1 {
	// 	t = timestamps[0].T
	// } else if len(timestamps) > 1 {
	// 	return &t, errors.New("more than one timestamp in db")
	// }

	// return &t, nil
	return nil, nil
}

func (sdb *sqlDB) UpdateTimestamp(t *time.Time) error {
	// var ts timestamp.Timestamp
	// err := sdb.db.FirstOrCreate(&ts, &timestamp.Timestamp{}).Error

	// if err != nil {
	// 	return errors.Wrap(err, "failed to get timestamp")
	// }

	// ts.T = *t

	// err = sdb.db.Save(&ts).Error
	// return errors.Wrap(err, "failed to save timestamp")
	return nil
}

func (sdb *sqlDB) AllFeeds() ([]*feed.Feed, error) {
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

func (sdb *sqlDB) MatchingFeed(f *feed.Feed) (*feed.Feed, error) {
	// var feed feed.Feed
	// err := sdb.db.First(&feed, f).Error
	// return &feed, errors.Wrap(err, "failed to get feed")
	return nil, nil
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

func (sdb *sqlDB) SaveFeed(f *feed.Feed) error {
	// var existingFeed feed.Feed
	// err := sdb.db.FirstOrCreate(&existingFeed, f).Error
	// return errors.Wrap(err, "failed to save feed")
	return nil
}

func (sdb *sqlDB) AllTags() ([]*feed.Tag, error) {
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
	// var tag feed.Tag
	// err := sdb.db.First(&tag, t).Error
	// return &tag, errors.Wrap(err, "failed to get tag")
	return nil, nil
}

func (sdb *sqlDB) SaveTagToFeed(t *feed.Tag, f *feed.Feed) error {
	// var existingFeed feed.Feed
	// err := sdb.db.Find(&existingFeed, f).Error
	// if err != nil {
	// 	return errors.Wrap(err, "failed to get feed")
	// }

	// t.FeedID = existingFeed.ID

	// var existingTag feed.Tag
	// err = sdb.db.FirstOrCreate(&existingTag, t).Error
	// return errors.Wrap(err, "failed to save tag")
	return nil
}

func (sdb *sqlDB) AllItems() ([]*feed.Item, error) {
	var items []*feed.Item
	rows, err := sdb.db.Query("select id, name, email, title, description, link, published, feed_id from items;")
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
	// var items []*feed.Item
	// err := sdb.db.Find(&items, i).Error
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to get matching item")
	// }
	// if len(items) > 0 {
	// 	return items[0], nil
	// }

	// return &feed.Item{}, nil
	return nil, nil
}

func (sdb *sqlDB) ItemsFromFeed(f *feed.Feed) ([]*feed.Item, error) {
	var items []*feed.Item

	stmt, err := sdb.db.Prepare("select id, name, email, title, description, link, published, feed_id from items where feed_id=?;")
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

func (sdb *sqlDB) SaveItemToFeed(i *feed.Item, f *feed.Feed) error {
	// var existingFeed feed.Feed
	// err := sdb.db.Find(&existingFeed, f).Error
	// if err != nil {
	// 	return errors.Wrap(err, "failed to get feed")
	// }

	// i.FeedID = existingFeed.ID

	// return sdb.SaveItem(i)
	return nil
}

func (sdb *sqlDB) SaveItem(i *feed.Item) error {
	// err := sdb.db.Save(i).Error
	// return errors.Wrap(err, "failed to save item")
	return nil
}

func (sdb *sqlDB) UpdateItem(i *feed.Item) error {
	// err := sdb.db.Model(&feed.Item{}).Update(i).Error
	// return errors.Wrap(err, "failed to update item")
	return nil
}

func (sdb *sqlDB) Close() error {
	err := sdb.db.Close()
	return errors.Wrap(err, "unable to close db")
}
