package list

import (
	"gonews/db"
	gndb "gonews/db" // TODO: rename
	"sort"
)

type List struct {
	Title string
	Items []*gndb.Item
}

// TODO
func FromFeeds(feeds []*gndb.Feed) (*List, error) {
	return &List{}, nil
}

func New(db db.DB, title, tagName string) *List {
	var tags []*gndb.Tag
	if tagName != "" {
		db.Find(&tags, &gndb.Tag{
			Name: tagName,
		})
	}

	var feeds []*gndb.Feed
	if len(tags) > 0 {
		tag := tags[0]
		db.Model(tag).Related(&feeds)
	} else {
		db.Find(&feeds, &gndb.Feed{})
	}

	var items []*gndb.Item
	for _, f := range feeds {
		if f == nil {
			continue
		}

		var nextItems []*gndb.Item
		db.Model(f).Related(&nextItems)

		items = append(items, nextItems...)
	}

	l := &List{
		Title: title,
		Items: items,
	}
	sort.Sort(l)

	return l
}

func (l *List) Len() int {
	return len(l.Items)
}

func (l *List) Less(i, j int) bool {
	t1 := l.Items[i].Published
	t2 := l.Items[j].Published
	return t1.After(t2)
}

func (l *List) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}
