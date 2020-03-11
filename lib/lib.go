package lib

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
)

const (
	ByteLimit = 100
)

type FileInfo interface{}

type File interface {
	Close() error
	Read([]byte) (int, error)
	WriteString(string) (int, error)
}

type FS interface {
	Open(string) (File, error)
	OpenFile(string, int, os.FileMode) (File, error)
	Stat(string) (FileInfo, error)
}

type PersistentTimestamp interface {
	Parse(FS) (*time.Time, error)
	Update(FS, *time.Time) error
}

type TimestampFile struct {
	Path string
}

func (tf *TimestampFile) Parse(fs FS) (*time.Time, error) {
	var parsedTime time.Time
	if _, err := fs.Stat(tf.Path); os.IsNotExist(err) {
		return &parsedTime, nil
	} else {
		if f, err := fs.Open(tf.Path); err != nil {
			return nil, err
		} else {
			// TODO: use ReadAll from ioutil instead
			bytes := make([]byte, ByteLimit, ByteLimit)
			if byteCount, err := f.Read(bytes); err != nil {
				return nil, err
			} else {
				timestampString := string(bytes[:byteCount])
				if parsedTime, err := time.Parse(time.UnixDate, timestampString); err != nil {
					return nil, err
				} else {
					return &parsedTime, nil
				}
			}
		}

	}
}

func (tf *TimestampFile) Update(fs FS, t *time.Time) error {
	if f, err := fs.OpenFile(tf.Path, os.O_RDWR|os.O_CREATE, 0644); err != nil {
		return err
	} else {
		if _, err := f.WriteString(t.Format(time.UnixDate)); err != nil {
			return err
		} else {
			if err := f.Close(); err != nil {
				return err
			}
		}

		return nil
	}
}

type AppConfig interface {
	SetConfigName(string)
	AddConfigPath(string)
	ReadInConfig() error
	Get(string) interface{}
	GetString(string) string
}

type DB interface {
	FirstOrCreate(interface{}, ...interface{}) DB
	Find(interface{}, ...interface{}) DB
	Order(interface{}, ...bool) DB
	Close() error
	Model(interface{}) DB
	Update(...interface{}) DB
	Where(interface{}, ...interface{}) DB
}

type GofeedURLParser interface {
	ParseURL(string) (*gofeed.Feed, error)
}

type FeedItem struct {
	gorm.Model
	Title       string
	Description string
	Link        string
	Published   time.Time
	Url         string
	AuthorName  string
	AuthorEmail string
	Hide        bool
	Tags        string
}

// TODO: make this a gorm model, and a table in the db
type FeedConfig struct {
	Url  string
	Tags string
}

type FeedFetcher interface {
	FetchFeeds(AppConfig, GofeedURLParser, DB) error
}

// TODO: rename?
type DefaultFeedFetcher struct{}

func ConvertToFeedItem(goFeedItem *gofeed.Item, feedConfig *FeedConfig) *FeedItem {
	var goFeedAuthorName string
	var goFeedAuthorEmail string
	if goFeedItem.Author != nil {
		goFeedAuthorName = goFeedItem.Author.Name
		goFeedAuthorEmail = goFeedItem.Author.Email
	}

	return &FeedItem{
		Title:       goFeedItem.Title,
		Description: goFeedItem.Description,
		Link:        goFeedItem.Link,
		Published:   *goFeedItem.PublishedParsed,
		Url:         feedConfig.Url,
		AuthorName:  goFeedAuthorName,
		AuthorEmail: goFeedAuthorEmail,
		Tags:        feedConfig.Tags,
	}
}

func ProcessGoFeedItem(db DB, goFeedItem *gofeed.Item, feedConfig *FeedConfig) {
	fi := ConvertToFeedItem(goFeedItem, feedConfig)

	// If an existing item with the same title, link & feed url exists, do
	// not insert the item
	var matchingFeedItems []*FeedItem
	db.Find(&matchingFeedItems, &FeedItem{
		Title: fi.Title,
		Link:  fi.Link,
		Url:   feedConfig.Url,
	})
	if len(matchingFeedItems) > 0 {
		return
	}

	var existingFeedItem FeedItem
	db.FirstOrCreate(&existingFeedItem, fi)
}

func (dff *DefaultFeedFetcher) FetchFeeds(appConfig AppConfig, feedParser GofeedURLParser, db DB) error {
	feedConfigs := appConfig.Get("feeds").([]interface{})
	for _, feedConfig := range feedConfigs {
		feedConfigMap := feedConfig.(map[string]interface{})
		feedUrl, exists := feedConfigMap["url"].(string)
		if !exists {
			return errors.New("Feed config must contain url")
		}

		feedConfig := &FeedConfig{Url: feedUrl}

		if feedTags, exists := feedConfigMap["tags"].([]interface{}); exists {
			// TODO: see if there's a 'map' function to do this; if not, make one?
			feedTagsCount := len(feedTags)
			feedTagsStrings := make([]string, feedTagsCount, feedTagsCount)
			for i, v := range feedTags {
				feedTagsStrings[i] = fmt.Sprintf("<%v>", v.(string))
			}

			feedConfig.Tags = strings.Join(feedTagsStrings, ",")
		}

		if nextFeed, err := feedParser.ParseURL(feedUrl); err != nil {
			fmt.Printf("Warning: could not retrieve feed from %v\n", feedUrl)
		} else if len(nextFeed.Items) < 1 {
			fmt.Printf("Warning: %v feed is empty\n", feedUrl)
		} else {
			for _, goFeedItem := range nextFeed.Items {
				ProcessGoFeedItem(db, goFeedItem, feedConfig)
			}

		}
	}

	return nil
}

func FetchFeedsAfterDelay(appConfig AppConfig, fs FS, pt PersistentTimestamp, ff FeedFetcher, fp GofeedURLParser, db DB) error {
	if lastUpdatedTime, err := pt.Parse(fs); err != nil {
		return err
	} else {
		if d, err := time.ParseDuration(appConfig.GetString("feed_fetch_period")); err != nil {
			return err
		} else {
			feedFetchTime := lastUpdatedTime.Add(d)
			if delay, err := time.ParseDuration(appConfig.GetString("feed_fetch_delay")); err != nil {
				return err
			} else {
				for {
					if time.Now().After(feedFetchTime) {
						if err := ff.FetchFeeds(appConfig, fp, db); err != nil {
							return err
						}

						currentTime := time.Now()
						pt.Update(fs, &currentTime)

						return nil
					} else {
						time.Sleep(delay)
					}
				}
			}
		}
	}
}
