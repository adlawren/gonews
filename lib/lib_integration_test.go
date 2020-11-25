package lib

import (
	"fmt"
	"gonews/config"
	"gonews/db"
	"gonews/feed"
	"gonews/rss"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func testConfig(t *testing.T) (*config.Config, error) {
	d, err := time.ParseDuration("10s")
	if err != nil {
		return nil, err
	}

	return &config.Config{
		AppTitle: "Test Title",
		Feeds: []*config.FeedConfig{
			{
				URL: "http://localhost:8081",
				Tags: []string{
					"tag1",
				},
			},
		},
		FetchPeriod: d,
	}, nil
}

var testDBConfig *config.DBConfig = &config.DBConfig{
	DSN: fmt.Sprintf("file:/tmp/gonews/test/%d/db.sqlite3",
		time.Now().Unix()),
}

var expectedFeeds []*feed.Feed = []*feed.Feed{
	{
		URL: "http://localhost:8081",
	},
}

var expectedTags []*feed.Tag = []*feed.Tag{
	{
		Name:   "tag1",
		FeedID: 1,
	},
}

func expectedItems() []*feed.Item {
	pubDate := time.Date(2004, time.October, 19, 12, 0, 0, 0, time.UTC)
	return []*feed.Item{
		{
			Person: gofeed.Person{
				Name:  "",
				Email: "",
			},
			Title:       "RSS Solutions for Restaurants",
			Description: "RSS Solutions for Restaurants description",
			Link:        "http://www.feedforall.com/restaurant.htm",
			Published:   pubDate,
			Hide:        false,
			FeedID:      1,
		},
		{
			Person: gofeed.Person{
				Name:  "",
				Email: "",
			},
			Title:       "RSS Solutions for Schools and Colleges",
			Description: "RSS Solutions for Schools and Colleges description",
			Link:        "http://www.feedforall.com/schools.htm",
			Published:   pubDate,
			Hide:        false,
			FeedID:      1,
		},
		{
			Person: gofeed.Person{
				Name:  "",
				Email: "",
			},
			Title:       "RSS Solutions for Computer Service Companies",
			Description: "RSS Solutions for Computer Service Companies description",
			Link:        "http://www.feedforall.com/computer-service.htm",
			Published:   pubDate,
			Hide:        false,
			FeedID:      1,
		},
	}
}

func initDB(t *testing.T, dbCfg *config.DBConfig) (db.DB, error) {
	log.Info().Msgf("Initializing test DB: %s", dbCfg.DSN)
	if !strings.HasPrefix(dbCfg.DSN, "file:") {
		return nil, errors.New(fmt.Sprintf(
			"invalid DSN (%s), expected local file", dbCfg.DSN))
	}

	path := strings.TrimPrefix(dbCfg.DSN, "file:")
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(path), os.ModeDir)
		if err != nil {
			return nil, err
		}

		_, err = os.Create(path)
		if err != nil {
			return nil, err
		}
	}

	adb, err := db.New(dbCfg)
	if err != nil {
		return nil, err
	}

	err = adb.Migrate("../db/migrations")
	return adb, errors.Wrap(err, "failed to migrate DB")
}

func TestWatchFeeds(t *testing.T) {
	db, err := initDB(t, testDBConfig)
	assert.NoError(t, err)

	testCfg, err := testConfig(t)
	assert.NoError(t, err)

	err = InsertMissingFeeds(testCfg, db)
	assert.NoError(t, err)

	go func() {
		assert.NoError(t, rss.Serve("test/sample.xml", 8081))
	}()

	go func() {
		assert.NoError(t, WatchFeeds(testCfg, testDBConfig))
	}()

	d, err := time.ParseDuration("30s")
	assert.NoError(t, err)

	time.Sleep(d)

	feeds, err := db.Feeds()
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFeeds), len(feeds))
	for i := 0; i < len(feeds); i++ {
		assert.Equal(t, expectedFeeds[i].String(), feeds[i].String())
	}

	tags, err := db.Tags()
	assert.NoError(t, err)
	assert.Equal(t, len(expectedTags), len(tags))
	for i := 0; i < len(tags); i++ {
		assert.Equal(t, expectedTags[i].String(), tags[i].String())
	}

	feeds, err = db.Feeds()
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFeeds), len(feeds))
	for i := 0; i < len(feeds); i++ {
		assert.Equal(t, expectedFeeds[i].String(), feeds[i].String())
	}

	feed, err := db.MatchingFeed(expectedFeeds[0])
	assert.NoError(t, err)
	items, err := db.ItemsFromFeed(feed)
	expectedItems := expectedItems()
	assert.Equal(t, len(items), len(expectedItems))
	for i := 0; i < len(items); i++ {
		assert.Equal(t, items[i].String(), expectedItems[i].String())
	}
}
