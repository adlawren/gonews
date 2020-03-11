package page

import (
	"gonews/db"
	"gonews/feed"
	"sort"

	"github.com/pkg/errors"
)

// Page contains the data needed to render a web page
type Page struct {
	Title string
	Items []*feed.Item
}

// New creates a new page
func New(db db.DB, title, tagName string) (*Page, error) {
	var err error
	p := &Page{
		Title: title,
	}

	var tag *feed.Tag
	if tagName != "" {
		tag, err = db.MatchingTag(&feed.Tag{Name: tagName})
	}

	if err != nil {
		return p, errors.Wrap(err, "failed to get tag")
	}

	var feeds []*feed.Feed
	if tag != nil {
		feeds, err = db.FeedsFromTag(tag)
	} else {
		feeds, err = db.AllFeeds()
	}

	if err != nil {
		return p, errors.Wrap(err, "failed to get feeds")
	}

	var items []*feed.Item
	for _, f := range feeds {
		if f == nil {
			continue
		}

		nextItems, err := db.ItemsFromFeed(f)
		if err != nil {
			return p, errors.Wrap(err, "failed to get items from feed")
		}

		items = append(items, nextItems...)
	}

	p.Items = items
	sort.Sort(p)

	return p, nil
}

// TODO: remove these methods?
// They're needed for sorting, but might be able to 'ORDER by PUBLISHED' instead

// Len returns the number of feed items in the page
func (p *Page) Len() int {
	return len(p.Items)
}

// Less returns a bool indicating whether the publish date of the feed item at
// the first position is earlier than that of the feed item at the second
// position - true if so, and false otherwise
func (p *Page) Less(i, j int) bool {
	t1 := p.Items[i].Published
	t2 := p.Items[j].Published
	return t1.After(t2)
}

// Swap replaces the feed item at the first position with the feed item at the
// second position, and vice versa
func (p *Page) Swap(i, j int) {
	p.Items[i], p.Items[j] = p.Items[j], p.Items[i]
}
