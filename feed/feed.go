package feed

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
)

// Feed contains the data associated with a feed stored in the database
type Feed struct {
	gorm.Model
	URL string
}

// Tag contains the data associated with a feed tag stored in the database
// Tag lists could be serialized and stored as strings in feeds table instead,
// but this seems cleaner
type Tag struct {
	gorm.Model
	Name   string
	FeedID uint
}

// Item contains the data associated with a feed item stored in the database
// Some fields copied from gofeed.Item; couldn't embed gofeed.Item because it
// includes slices, which can't be directly saved to the DB
type Item struct {
	gorm.Model
	gofeed.Person
	Title       string
	Description string
	Link        string
	Published   time.Time
	Hide        bool
	FeedID      uint
}

// FromGofeedItem overrides the fields in the item with those from the given
// gofeed item
func (i *Item) FromGofeedItem(gfi *gofeed.Item) error {
	if i == nil || gfi == nil {
		return errors.New("item pointer is nil")
	}

	var name string
	var email string
	if gfi.Author != nil {
		name = gfi.Author.Name
		email = gfi.Author.Email
	}

	var published time.Time
	if gfi.PublishedParsed != nil {
		published = *gfi.PublishedParsed
	}

	i.Person = gofeed.Person{
		Name:  name,
		Email: email,
	}
	i.Title = gfi.Title
	i.Description = gfi.Description
	i.Link = gfi.Link
	i.Published = published

	return nil
}
