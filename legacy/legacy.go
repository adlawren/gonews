package legacy

import (
	"time"

	"github.com/jinzhu/gorm"
)

// FeedItem is deprecated and unused;
// preserved for db migrations
type FeedItem struct {
	gorm.Model
	Title       string
	Description string
	Link        string
	Published   time.Time
	Url         string // TODO: change to uppercase
	AuthorName  string
	AuthorEmail string
	Hide        bool
}
