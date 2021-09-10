package feed

import (
	"fmt"
	"html"
	"time"

	"github.com/mmcdole/gofeed"
)

// Feed contains the data associated with a feed stored in the database
type Feed struct {
	ID         uint
	URL        string
	FetchLimit uint
}

func (f Feed) String() string {
	return fmt.Sprintf("Feed{URL: %s, FetchLimit: %d}", f.URL, f.FetchLimit)
}

// Tag contains the data associated with a feed tag stored in the database
// Tag lists could be serialized and stored as strings in feeds table instead,
// but this seems cleaner
type Tag struct {
	ID     uint
	Name   string
	FeedID uint
}

func (t Tag) String() string {
	return fmt.Sprintf("Tag{Name: %s, FeedID: %d}", t.Name, t.FeedID)
}

// Item contains the data associated with a feed item stored in the database
// Some fields copied from gofeed.Item; couldn't embed gofeed.Item because it
// includes slices, which can't be directly saved to the DB
type Item struct {
	ID          uint
	Name        string
	Email       string
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Published   time.Time `json:"time"`
	Hide        bool      `json:"hide"`
	FeedID      uint      `json:"feed_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func (i Item) String() string {
	return fmt.Sprintf(
		"Item{Author name: %s, "+
			"Author email: %s, "+
			"Title: %s, "+
			"Description: %s, "+
			"Link: %s, "+
			"Published: %s, "+
			"Hide: %t, "+
			"FeedID: %d}",
		i.Name,
		i.Email,
		i.Title,
		i.Description,
		i.Link,
		i.Published,
		i.Hide,
		i.FeedID)
}

// FromGofeedItem overrides the fields in the item with those from the given
// gofeed item
func (i *Item) FromGofeedItem(gfi *gofeed.Item) error {
	if i == nil || gfi == nil {
		return fmt.Errorf("item pointer is nil")
	}

	var name string
	var email string
	if gfi.Author != nil {
		name = html.EscapeString(gfi.Author.Name)
		email = html.EscapeString(gfi.Author.Email)
	}

	var published time.Time
	if gfi.PublishedParsed != nil {
		published = *gfi.PublishedParsed
	}

	i.Name = name
	i.Email = email
	i.Title = html.EscapeString(gfi.Title)
	i.Description = html.EscapeString(gfi.Description)
	i.Link = html.EscapeString(gfi.Link)
	i.Published = published

	return nil
}
