package main

import (
	"fmt"
	"gonews/config"
	"gonews/db"
	"gonews/feed"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
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
				URL: "http://localhost:8080",
				Tags: []string{
					"tag1",
				},
			},
		},
		FetchPeriod: d,
	}, nil
}

var testDBConfig *config.DBConfig = &config.DBConfig{
	Path: fmt.Sprintf("/tmp/gonews/test/%d/db.sqlite3",
		time.Now().Unix()),
	DSN: fmt.Sprintf("file:/tmp/gonews/test/%d/db.sqlite3",
		time.Now().Unix()),
}

var expectedFeeds []*feed.Feed = []*feed.Feed{
	{
		URL: "http://localhost:8080",
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
	_, err := os.Stat(dbCfg.Path)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(dbCfg.Path), os.ModeDir)
		if err != nil {
			return nil, err
		}

		_, err = os.Create(dbCfg.Path)
		if err != nil {
			return nil, err
		}
	}

	adb, err := db.New(dbCfg)
	if err != nil {
		return nil, err
	}

	return adb, db.Migrate(adb)
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	// Source: https://www.feedforall.com/sample.xml
	http.ServeFile(w, r, "assets/sample.xml")
}

func serveRSS(t *testing.T) {
	http.HandleFunc("/", rssHandler)

	err := http.ListenAndServe(":8080", nil)
	assert.NoError(t, err)
}

func TestWatchFeeds(t *testing.T) {
	db, err := initDB(t, testDBConfig)
	assert.NoError(t, err)

	testCfg, err := testConfig(t)
	assert.NoError(t, err)

	go serveRSS(t)
	go watchFeeds(testCfg, testDBConfig)

	d, err := time.ParseDuration("30s")
	assert.NoError(t, err)

	time.Sleep(d)

	feeds, err := db.AllFeeds()
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFeeds), len(feeds))
	for i := 0; i < len(feeds); i++ {
		assert.Equal(t, expectedFeeds[i].String(), feeds[i].String())
	}

	tag, err := db.MatchingTag(expectedTags[0])
	feeds, err = db.FeedsFromTag(tag)
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
