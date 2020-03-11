package main

import (
	"fmt"
	"gonews/config"
	"gonews/feed"
	"gonews/mock_db"
	"gonews/mock_parser"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestFetchFeedsReturnsErrorWhenFeedSaveFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockErr := mockError()

	parser := mock_parser.NewMockGofeedParser(ctrl)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().SaveFeed(gomock.Any()).Return(mockErr)

	err := fetchFeeds(mockCfg, parser, db)
	expectedErrMsg := fmt.Sprintf(
		"failed to save feed: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestFetchFeedsReturnsErrorWhenTagSaveFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockErr := mockError()

	parser := mock_parser.NewMockGofeedParser(ctrl)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(mockErr)

	err := fetchFeeds(mockCfg, parser, db)
	expectedErrMsg := fmt.Sprintf(
		"failed to save tag: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestFetchFeedsReturnsErrorWhenFeedParseFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockErr := mockError()

	parser := mock_parser.NewMockGofeedParser(ctrl)
	parser.EXPECT().ParseURL(mockCfg.Feeds[0].URL).Return(nil, mockErr)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)

	err := fetchFeeds(mockCfg, parser, db)
	expectedErrMsg := fmt.Sprintf(
		"failed to parse feed: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestFetchFeedsReturnsErrorWhenItemSaveFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockGofeed := mockGofeed()
	mockErr := mockError()

	parser := mock_parser.NewMockGofeedParser(ctrl)
	parser.EXPECT().ParseURL(mockCfg.Feeds[0].URL).Return(mockGofeed, nil)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(&feed.Item{}, nil)
	db.EXPECT().SaveItemToFeed(gomock.Any(), gomock.Any()).Return(mockErr)

	err := fetchFeeds(mockCfg, parser, db)
	expectedErrMsg := fmt.Sprintf(
		"failed to save item: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestFetchFeedsReturnsErrorWhenItemInitFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockGofeed := &gofeed.Feed{
		Items: []*gofeed.Item{nil},
	}

	parser := mock_parser.NewMockGofeedParser(ctrl)
	parser.EXPECT().ParseURL(mockCfg.Feeds[0].URL).Return(mockGofeed, nil)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)

	err := fetchFeeds(mockCfg, parser, db)
	assert.EqualError(
		t,
		err,
		"failed to initialize item: item pointer is nil")
}

func TestFetchFeedsReturnsErrorWhenMatchingItemFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockGofeed := mockGofeed()
	mockErr := mockError()

	parser := mock_parser.NewMockGofeedParser(ctrl)
	parser.EXPECT().ParseURL(mockCfg.Feeds[0].URL).Return(mockGofeed, nil)
	parser.EXPECT().ParseURL(mockCfg.Feeds[1].URL).Return(mockGofeed, nil)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(nil, mockErr)

	err := fetchFeeds(mockCfg, parser, db)
	expectedErrMsg := fmt.Sprintf(
		"failed to get matching item: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestFetchFeedsSavesItems(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockGofeed := mockGofeed()
	item := &feed.Item{}

	parser := mock_parser.NewMockGofeedParser(ctrl)
	parser.EXPECT().ParseURL(mockCfg.Feeds[0].URL).Return(mockGofeed, nil)
	parser.EXPECT().ParseURL(mockCfg.Feeds[1].URL).Return(mockGofeed, nil)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(item, nil)
	db.EXPECT().SaveItemToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(item, nil)
	db.EXPECT().SaveItemToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(item, nil)
	db.EXPECT().SaveItemToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(item, nil)
	db.EXPECT().SaveItemToFeed(gomock.Any(), gomock.Any()).Return(nil)

	err := fetchFeeds(mockCfg, parser, db)
	assert.NoError(t, err)
}

func TestFetchFeedsSkipsMatchingItems(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockGofeed := mockGofeed()
	item1 := mockGofeed.Items[0]
	item2 := mockGofeed.Items[1]

	parser := mock_parser.NewMockGofeedParser(ctrl)
	parser.EXPECT().ParseURL(mockCfg.Feeds[0].URL).Return(mockGofeed, nil)
	parser.EXPECT().ParseURL(mockCfg.Feeds[1].URL).Return(mockGofeed, nil)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(
		&feed.Item{Title: item1.Title}, nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(
		&feed.Item{Title: item2.Title}, nil)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().SaveTagToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(&feed.Item{}, nil)
	db.EXPECT().SaveItemToFeed(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(&feed.Item{}, nil)
	db.EXPECT().SaveItemToFeed(gomock.Any(), gomock.Any()).Return(nil)

	err := fetchFeeds(mockCfg, parser, db)
	assert.NoError(t, err)
}

func randTag() string {
	return fmt.Sprintf("tag %v", rand.Int())
}

func randTags(n int) []string {
	tags := make([]string, n, n)
	for i := 0; i < len(tags); i++ {
		tags[i] = randTag()
	}

	return tags
}

func randFeedConfig() *config.FeedConfig {
	return &config.FeedConfig{
		URL:  fmt.Sprintf("test url %d", rand.Int()),
		Tags: randTags(2),
	}
}

func randFeedConfigs(n int) []*config.FeedConfig {
	feeds := make([]*config.FeedConfig, n, n)
	for i := 0; i < len(feeds); i++ {
		feeds[i] = randFeedConfig()
	}

	return feeds
}

func mockConfig(t *testing.T) *config.Config {
	d, err := time.ParseDuration("10m")
	assert.NoError(t, err, "Failed to parse mock duration")

	feeds := randFeedConfigs(2)
	return &config.Config{
		AppTitle:    "Test Title",
		Feeds:       feeds,
		FetchPeriod: d,
	}
}

func randGofeedItem() *gofeed.Item {
	now := time.Now()
	return &gofeed.Item{
		Title:           fmt.Sprintf("Title %d", rand.Int()),
		Description:     fmt.Sprintf("Description %d", rand.Int()),
		Link:            fmt.Sprintf("Link %d", rand.Int()),
		PublishedParsed: &now,
		Author: &gofeed.Person{
			Name:  fmt.Sprintf("Name %d", rand.Int()),
			Email: fmt.Sprintf("Email %d", rand.Int()),
		},
	}
}

func randGofeedItems(n int) []*gofeed.Item {
	items := make([]*gofeed.Item, n, n)
	for i := 0; i < len(items); i++ {
		items[i] = randGofeedItem()
	}
	return items
}

func mockGofeed() *gofeed.Feed {
	return &gofeed.Feed{
		Items: randGofeedItems(2),
	}
}

func mockError() error {
	return errors.New("mock error")
}
