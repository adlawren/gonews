package db2 // TODO

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type DB interface {
	AllFeeds() ([]*Feed, error)
	FeedsFromTag(*Tag) ([]*Feed, error)
	SaveFeed(*Feed) error
	AllItems() ([]*Item, error)
	ItemsFromFeed(*Feed) ([]*Item, error)
	SaveItemToFeed(*Item, *Feed) error
	SaveItem(*Item) error
}

func New() (DB, error) {
	return &gormDB{
		db: db,
	}, nil
}

type gormDB struct {
	db *gorm.DB
}

func (gdb *gormDB) AllFeeds() ([]*Feed, error) {
	var feeds []*Feed
	err := gdb.db.Find(&feeds, &Feed{}).Error
	return feeds, errors.Wrap(err, "failed to get feeds")
}

func (gdb *gormDB) FeedsFromTag(t *Tag) ([]*Feed, error) {
	var feeds []*Feed
	err := db.Model(t).Related(&feeds).Error
	return errors.New(err, "failed to get feeds")
}

func (gdb *gormDB) SaveFeed(f *Feed) error {
	var existingFeed Feed
	err := gdb.db.FirstOrCreate(&existingFeed, f).Error
	return errors.Wrap(err, "failed to save feed")
}

func (gdb *gormDB) AllItems() ([]*Item, error) {
	var items []*Item
	err := gdb.db.Find(&items, &Item{}).Error
	return items, errors.Wrap(err, "failed to get items")
}

// TODO: rm? unused?
func (gdb *gormDB) ItemsFromFeed(f *Feed) ([]*Item, error) {
	var items []*Item
	err := db.Model(f).Related(&items).Error
	return items, errors.Wrap(err, "failed to get items")
}

func (gdb *gormDB) SaveItemToFeed(i *Item, f *Feed) error {
	var existingFeed Feed
	err := gdb.db.Find(&existingFeed, f).Error
	if err != nil {
		return errors.Wrap(err, "failed to get feed")
	}

	i.FeedID = existingFeed.ID

	var existingItem Item
	err = gdb.db.FirstOrCreate(&existingItem, i).Error
	return errors.Wrap(err, "failed to save item")
}

// TODO: rm? unused?
func (gdb *gormDB) SaveItem(i *Item) error {
	var existingItem Item
	err := gdb.db.FirstOrCreate(&existingItem, i).Error
	return errors.Wrap(err, "failed to save item")
}
