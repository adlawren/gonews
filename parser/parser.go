package parser

import "github.com/mmcdole/gofeed"

// GofeedParser contains the method needed to parse a gofeed.Feed from a given
// URL
type GofeedParser interface {
	ParseURL(string) (*gofeed.Feed, error)
}

// New creates a new instance of a struct compatible with the GofeedParser
// interface
func New() (GofeedParser, error) {
	return gofeed.NewParser(), nil
}
