package db

import (
	"gonews/config"
	"gonews/feed"
	"gonews/timestamp"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// DB contains the methods needed to store and read data from the underlying
// database
type DB interface {
	Timestamp() (*time.Time, error)
	UpdateTimestamp(*time.Time) error
	AllFeeds() ([]*feed.Feed, error) // TODO: rm?
	MatchingFeed(*feed.Feed) (*feed.Feed, error)
	FeedsFromTag(*feed.Tag) ([]*feed.Feed, error)
	SaveFeed(*feed.Feed) error
	MatchingTag(*feed.Tag) (*feed.Tag, error)
	SaveTagToFeed(*feed.Tag, *feed.Feed) error
	AllItems() ([]*feed.Item, error)
	MatchingItem(*feed.Item) (*feed.Item, error)
	ItemsFromFeed(*feed.Feed) ([]*feed.Item, error)
	SaveItemToFeed(*feed.Item, *feed.Feed) error
	SaveItem(*feed.Item) error
	Close() error
}

// New creates a struct which supports the operations in the DB interface
func New(cfg *config.DBConfig) (DB, error) {
	db, err := gorm.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open db")
	}

	return &gormDB{db: db}, nil
}

type gormDB struct {
	db *gorm.DB
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

func (gdb *gormDB) Close() error {
	err := gdb.db.Close()
	return errors.Wrap(err, "unable to close db")
}
