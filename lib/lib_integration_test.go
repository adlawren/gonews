package lib

import (
	"context"
	"gonews/config"
	"gonews/db/orm/query/clause"
	"gonews/feed"
	"gonews/rss"
	"gonews/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	migrationsDir = "../db/migrations"
)

func testConfig(t *testing.T) *config.Config {
	autoDismissAfter, err := time.ParseDuration("1s")
	assert.NoError(t, err)

	autoDismissPeriod, err := time.ParseDuration("1s")
	assert.NoError(t, err)

	fetchPeriod, err := time.ParseDuration("1s")
	assert.NoError(t, err)

	return &config.Config{
		AppTitle: "Test Title",
		Feeds: []*config.FeedConfig{
			{
				URL: "http://localhost:8081",
				Tags: []string{
					"tag1",
				},
				AutoDismissAfter: autoDismissAfter,
			},
		},
		FetchPeriod:       fetchPeriod,
		AutoDismissPeriod: autoDismissPeriod,
	}
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
			Name:        "",
			Email:       "",
			Title:       "RSS Solutions for Restaurants",
			Description: "RSS Solutions for Restaurants description",
			Link:        "http://www.feedforall.com/restaurant.htm",
			Published:   pubDate,
			Hide:        false,
			FeedID:      1,
		},
		{
			Name:        "",
			Email:       "",
			Title:       "RSS Solutions for Schools and Colleges",
			Description: "RSS Solutions for Schools and Colleges description",
			Link:        "http://www.feedforall.com/schools.htm",
			Published:   pubDate,
			Hide:        false,
			FeedID:      1,
		},
		{
			Name:        "",
			Email:       "",
			Title:       "RSS Solutions for Computer Service Companies",
			Description: "RSS Solutions for Computer Service Companies description",
			Link:        "http://www.feedforall.com/computer-service.htm",
			Published:   pubDate,
			Hide:        false,
			FeedID:      1,
		},
	}
}

func TestWatchFeeds(t *testing.T) {
	dbCfg, db := test.InitDB(t, migrationsDir)
	testCfg := testConfig(t)

	err := InsertMissingFeeds(testCfg, db)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := rss.Serve(ctx, "test/sample.xml", 8081)
		if ctx.Err() != context.Canceled {
			assert.NoError(t, err)
		}
	}()

	go func() {
		err := WatchFeeds(ctx, testCfg, dbCfg)
		if ctx.Err() != context.Canceled {
			assert.NoError(t, err)
		}
	}()

	d, err := time.ParseDuration("3s")
	assert.NoError(t, err)

	time.Sleep(d)

	var feeds []*feed.Feed
	err = db.All(&feeds)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFeeds), len(feeds))
	for i := 0; i < len(feeds); i++ {
		assert.Equal(t, expectedFeeds[i].String(), feeds[i].String())
	}

	var tags []*feed.Tag
	err = db.All(&tags)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedTags), len(tags))
	for i := 0; i < len(tags); i++ {
		assert.Equal(t, expectedTags[i].String(), tags[i].String())
	}

	var f feed.Feed
	err = db.Find(&f, clause.New("where url = ?", expectedFeeds[0].URL))
	assert.NoError(t, err)

	var items []*feed.Item
	err = db.FindAll(&items, clause.New("where feed_id = ?", f.ID))
	expectedItems := expectedItems()
	assert.Equal(t, len(items), len(expectedItems))
	for i := 0; i < len(items); i++ {
		assert.Equal(t, items[i].String(), expectedItems[i].String())
	}
}

func TestAutoDismissItems(t *testing.T) {
	dbCfg, db := test.InitDB(t, migrationsDir)
	testCfg := testConfig(t)

	err := InsertMissingFeeds(testCfg, db)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := rss.Serve(ctx, "test/sample.xml", 8081)
		if ctx.Err() != context.Canceled {
			assert.NoError(t, err)
		}
	}()

	go func() {
		err := WatchFeeds(ctx, testCfg, dbCfg)
		if ctx.Err() != context.Canceled {
			assert.NoError(t, err)
		}
	}()

	go func() {
		err := AutoDismissItems(ctx, testCfg, dbCfg)
		if ctx.Err() != context.Canceled {
			assert.NoError(t, err)
		}
	}()

	d, err := time.ParseDuration("3s")
	assert.NoError(t, err)

	time.Sleep(d)

	var f feed.Feed
	err = db.Find(&f, clause.New("where url = ?", testCfg.Feeds[0].URL))
	assert.NoError(t, err)

	var items []*feed.Item
	err = db.FindAll(&items, clause.New("where feed_id = ?", f.ID))
	assert.NoError(t, err)
	for _, item := range items {
		assert.True(t, item.Hide)
	}
}

func TestAutoDismissItemsIgnoresItemsYoungerThanAutoDismissAfter(t *testing.T) {
	dbCfg, db := test.InitDB(t, migrationsDir)
	testCfg := testConfig(t)

	autoDismissAfter, err := time.ParseDuration("1h")
	assert.NoError(t, err)

	testCfg.Feeds[0].AutoDismissAfter = autoDismissAfter

	err = InsertMissingFeeds(testCfg, db)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := rss.Serve(ctx, "test/sample.xml", 8081)
		if ctx.Err() != context.Canceled {
			assert.NoError(t, err)
		}
	}()

	go func() {
		err := WatchFeeds(ctx, testCfg, dbCfg)
		if ctx.Err() != context.Canceled {
			assert.NoError(t, err)
		}
	}()

	go func() {
		err := AutoDismissItems(ctx, testCfg, dbCfg)
		if ctx.Err() != context.Canceled {
			assert.NoError(t, err)
		}
	}()

	d, err := time.ParseDuration("3s")
	assert.NoError(t, err)

	time.Sleep(d)

	var f feed.Feed
	err = db.Find(&f, clause.New("where url = ?", testCfg.Feeds[0].URL))
	assert.NoError(t, err)

	var items []*feed.Item
	err = db.FindAll(&items, clause.New("where feed_id = ?", f.ID))
	assert.NoError(t, err)
	for _, item := range items {
		assert.False(t, item.Hide)
	}
}
