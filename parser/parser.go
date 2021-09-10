package parser

import (
	"fmt"
	"gonews/feed"

	"github.com/mmcdole/gofeed"
)

// Parser contains the method needed to parse a list of items from a given RSS
// URL
type Parser interface {
	ParseURL(string) ([]*feed.Item, error)
}

// New creates a new instance of a struct compatible with the Parser interface
func New() (Parser, error) {
	return &gfParser{
		parser: gofeed.NewParser(),
	}, nil
}

type gfParser struct {
	parser *gofeed.Parser
}

func (p *gfParser) ParseURL(feedURL string) ([]*feed.Item, error) {
	var items []*feed.Item

	gfeed, err := p.parser.ParseURL(feedURL)
	if err != nil {
		return items, fmt.Errorf("failed to parse feed: %w", err)
	}

	for _, gitem := range gfeed.Items {
		var i feed.Item
		err := i.FromGofeedItem(gitem)
		if err != nil {
			return items, fmt.Errorf("failed to initialize item: %w", err)
		}

		items = append(items, &i)
	}

	return items, nil
}
