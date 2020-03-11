package lib

import (
	"fmt"
	"gonews/config"
	gndb "gonews/db" // TODO: rename
	"gonews/fs"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
)

const (
	byteLimit = 100
)

type PersistentTimestamp interface {
	Parse(fs.FS) (*time.Time, error) // TODO: rename? Time()
	Update(fs.FS, *time.Time) error  // TODO: rename UpdateTime
}

type TimestampFile struct {
	Path string
}

func (tf *TimestampFile) Parse(fs fs.FS) (*time.Time, error) {
	var parsedTime time.Time

	_, err := fs.Stat(tf.Path)
	if os.IsNotExist(err) {
		return &parsedTime, nil
	}

	f, err := fs.Open(tf.Path)
	if err != nil {
		return nil, err
	}

	// TODO: use ReadAll from ioutil instead
	bytes := make([]byte, byteLimit, byteLimit)
	byteCount, err := f.Read(bytes)
	if err != nil {
		return nil, err
	}

	timestampString := string(bytes[:byteCount])
	parsedTime, err = time.Parse(time.UnixDate, timestampString)
	if err != nil {
		return nil, err
	}

	return &parsedTime, nil
}

func (tf *TimestampFile) Update(fs fs.FS, t *time.Time) error {
	f, err := fs.OpenFile(tf.Path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	// TODO: maybe just `touch` the file and use the timestamp instead?
	_, err = f.WriteString(t.Format(time.UnixDate))
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

// TODO: rename
type GofeedURLParser interface {
	ParseURL(string) (*gofeed.Feed, error)
}

// TODO: rename
type FeedFetcher interface {
	FetchFeeds(config.Config, GofeedURLParser, gndb.DB) error
}

// TODO: rename?
type DefaultFeedFetcher struct{}

func ProcessGoFeedItem(db gndb.DB, gfi *gofeed.Item, cfg *config.FeedConfig) {
	//fi := gndb.FromGofeedItem(gfi)
	var fi gndb.Item
	fi.FromGofeedItem(gfi)

	// TODO: don't? Already created earlier..
	var existingFeed gndb.Feed
	db.FirstOrCreate(&existingFeed, &gndb.Feed{URL: cfg.URL})

	fi.FeedID = existingFeed.ID

	var existingFeedItem gndb.Item
	//db.FirstOrCreate(&existingFeedItem, fi)
	db.FirstOrCreate(&existingFeedItem, &fi)
}

func (dff *DefaultFeedFetcher) FetchFeeds(appConfig config.Config, feedParser GofeedURLParser, db gndb.DB) error {
	feedConfigs := appConfig.Get("feeds").([]interface{})
	for _, feedConfig := range feedConfigs {
		feedConfigMap := feedConfig.(map[string]interface{})
		feedURL, exists := feedConfigMap["url"].(string)
		if !exists {
			return errors.New("Feed config must contain url")
		}

		cfg := &config.FeedConfig{URL: feedURL}

		var existingFeed gndb.Feed
		db.FirstOrCreate(&existingFeed, &gndb.Feed{
			URL: feedURL,
		})

		if feedTags, exists := feedConfigMap["tags"].([]interface{}); exists {
			for _, ft := range feedTags {
				tagName, ok := ft.(string)
				if !ok {
					return errors.Errorf(
						"tag type is: %T, expected string",
						ft)
				}

				var existingTag gndb.Tag
				db.FirstOrCreate(&existingTag, &gndb.Tag{
					Name:   tagName,
					FeedID: existingFeed.ID,
				})
			}
		}

		if nextFeed, err := feedParser.ParseURL(feedURL); err != nil {
			fmt.Printf("Warning: could not retrieve feed from %v\n", feedURL)
		} else if len(nextFeed.Items) < 1 {
			fmt.Printf("Warning: %v feed is empty\n", feedURL)
		} else {
			for _, goFeedItem := range nextFeed.Items {
				ProcessGoFeedItem(db, goFeedItem, cfg)
			}
		}
	}

	return nil
}

// TODO: instead use time.Timer, select
func FetchFeedsAfterDelay(appConfig config.Config, fs fs.FS, pt PersistentTimestamp, ff FeedFetcher, fp GofeedURLParser, db gndb.DB) error {
	lastUpdatedTime, err := pt.Parse(fs)
	if err != nil {
		return err
	}

	d, err := time.ParseDuration(appConfig.GetString("feed_fetch_period"))
	if err != nil {
		return err
	}

	feedFetchTime := lastUpdatedTime.Add(d)
	delay, err := time.ParseDuration(appConfig.GetString("feed_fetch_delay"))
	if err != nil {
		return err
	}

	for {
		if time.Now().After(feedFetchTime) {
			if err := ff.FetchFeeds(appConfig, fp, db); err != nil {
				return err
			}

			currentTime := time.Now()
			pt.Update(fs, &currentTime)

			return nil
		}

		time.Sleep(delay)
	}
}
