package lib

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
	"os"
	"time"
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
			} else {
				return nil
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
}

type FeedParser interface {
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
}

type FeedFetcher interface {
	FetchFeeds(AppConfig, FeedParser, DB) error
}

// TODO: rename?
type DefaultFeedFetcher struct{}

func ConvertToFeedItem(feedUrl string, goFeedItem *gofeed.Item) *FeedItem {
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
		Url:         feedUrl,
		AuthorName:  goFeedAuthorName,
		AuthorEmail: goFeedAuthorEmail,
	}
}

// TODO: see if there's a way to avoid the redundancy - the feedUrl is in the feedConfigMap too
func ProcessGoFeedItem(db DB, goFeedItem *gofeed.Item, feedUrl string, feedConfigMap map[string]interface{}) {
	fi := ConvertToFeedItem(feedUrl, goFeedItem)

	var existingFeedItem FeedItem
	db.FirstOrCreate(&existingFeedItem, fi)
}

// TODO: revisit the commented code; may re-add the feature
func (dff *DefaultFeedFetcher) FetchFeeds(appConfig AppConfig, feedParser FeedParser, db DB) error {
	feedConfigs := appConfig.Get("feeds").([]interface{})
	for _, feedConfig := range feedConfigs {
		feedConfigMap := feedConfig.(map[string]interface{})
		feedUrl, exists := feedConfigMap["url"].(string)
		if !exists {
			return errors.New("Feed config must contain url")
		}

		//feedItemLimitInterface, exists := feedConfigMap["item_limit"]

		// var feedItemLimit int
		// if !exists {
		// 	feedItemLimit = int(appConfig.Get("item_limit").(int64))
		// } else {
		// 	feedItemLimit = int(feedItemLimitInterface.(int64))
		// }

		if nextFeed, err := feedParser.ParseURL(feedUrl); err != nil {
			fmt.Printf("Warning: could not retrieve feed from %v\n", feedUrl)
		} else if len(nextFeed.Items) < 1 {
			fmt.Printf("Warning: %v feed is empty\n", feedUrl)
		} else {
			// if len(nextFeed.Items) < feedItemLimit {
			// 	fmt.Printf("Warning: truncating item_limit; not enough items in %v feed\n", feedUrl)
			// 	feedItemLimit = len(nextFeed.Items)
			// }

			// for _, goFeedItem := range nextFeed.Items[:feedItemLimit] {
			// 	ProcessGoFeedItem(db, goFeedItem, feedUrl, feedConfigMap)
			// }

			for _, goFeedItem := range nextFeed.Items {
				ProcessGoFeedItem(db, goFeedItem, feedUrl, feedConfigMap)
			}

		}
	}

	return nil
}

func FetchFeedsAfterDelay(appConfig AppConfig, fs FS, pt PersistentTimestamp, ff FeedFetcher, fp FeedParser, db DB) error {
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
