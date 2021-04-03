package lib

import (
	"fmt"
	"gonews/config"
	"gonews/feed"
	"gonews/mock_db"
	"gonews/mock_parser"
	"gonews/test"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestInsertMissingFeedsReturnsErrorWhenMatchingFeedFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockErr := mockError()

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().MatchingFeed(gomock.Any()).Return(nil, mockErr)

	err := InsertMissingFeeds(mockCfg, db)
	expectedErrMsg := fmt.Sprintf(
		"failed to get matching feed: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestInsertMissingFeedsReturnsErrorWhenFeedSaveFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockErr := mockError()

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().MatchingFeed(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveFeed(gomock.Any()).Return(mockErr)

	err := InsertMissingFeeds(mockCfg, db)
	expectedErrMsg := fmt.Sprintf(
		"failed to save feed: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestInsertMissingFeedsReturnsErrorWhenMatchingTagFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockErr := mockError()

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().MatchingFeed(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().MatchingTag(gomock.Any()).Return(nil, mockErr)

	err := InsertMissingFeeds(mockCfg, db)
	expectedErrMsg := fmt.Sprintf(
		"failed to get matching tag: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestInsertMissingFeedsReturnsErrorWhenTagSaveFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCfg := mockConfig(t)
	mockErr := mockError()

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().MatchingFeed(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveFeed(gomock.Any()).Return(nil)
	db.EXPECT().MatchingTag(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveTag(gomock.Any()).Return(mockErr)

	err := InsertMissingFeeds(mockCfg, db)
	expectedErrMsg := fmt.Sprintf(
		"failed to save tag: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestFetchFeedsReturnsErrorWhenFeedsReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockErr := mockError()

	parser := mock_parser.NewMockParser(ctrl)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Feeds().Return(nil, mockErr)

	err := fetchFeeds(db, parser)
	expectedErrMsg := fmt.Sprintf(
		"failed to get feeds: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestFetchFeedsReturnsErrorWhenFeedParseFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockFeeds := mockFeeds()
	mockErr := mockError()

	parser := mock_parser.NewMockParser(ctrl)
	parser.EXPECT().ParseURL(mockFeeds[0].URL).Return(nil, mockErr)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Feeds().Return(mockFeeds, nil)

	err := fetchFeeds(db, parser)
	expectedErrMsg := fmt.Sprintf(
		"failed to parse feed: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestFetchFeedsReturnsErrorWhenItemSaveFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockFeeds := mockFeeds()
	mockFeedItems := test.MockItems()
	mockErr := mockError()

	parser := mock_parser.NewMockParser(ctrl)
	parser.EXPECT().ParseURL(mockFeeds[0].URL).Return(mockFeedItems, nil)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Feeds().Return(mockFeeds, nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveItem(gomock.Any()).Return(mockErr)

	err := fetchFeeds(db, parser)
	expectedErrMsg := fmt.Sprintf(
		"failed to save item: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestFetchFeedsReturnsErrorWhenMatchingItemFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockFeeds := mockFeeds()
	mockFeedItems := test.MockItems()
	mockErr := mockError()

	parser := mock_parser.NewMockParser(ctrl)
	parser.EXPECT().ParseURL(mockFeeds[0].URL).Return(mockFeedItems, nil)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Feeds().Return(mockFeeds, nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(nil, mockErr)

	err := fetchFeeds(db, parser)
	expectedErrMsg := fmt.Sprintf(
		"failed to get matching item: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestFetchFeedsSavesItems(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockFeeds := mockFeeds()
	mockFeedItems := test.MockItems()

	parser := mock_parser.NewMockParser(ctrl)
	parser.EXPECT().ParseURL(mockFeeds[0].URL).Return(mockFeedItems, nil)
	parser.EXPECT().ParseURL(mockFeeds[1].URL).Return(mockFeedItems, nil)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Feeds().Return(mockFeeds, nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveItem(gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveItem(gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveItem(gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveItem(gomock.Any()).Return(nil)

	err := fetchFeeds(db, parser)
	assert.NoError(t, err)
}

func TestFetchFeedsSkipsMatchingItems(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockFeeds := mockFeeds()
	mockFeedItems := test.MockItems()
	item1 := mockFeedItems[0]
	item2 := mockFeedItems[1]

	parser := mock_parser.NewMockParser(ctrl)
	parser.EXPECT().ParseURL(mockFeeds[0].URL).Return(mockFeedItems, nil)
	parser.EXPECT().ParseURL(mockFeeds[1].URL).Return(mockFeedItems, nil)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Feeds().Return(mockFeeds, nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(item1, nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(item2, nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveItem(gomock.Any()).Return(nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveItem(gomock.Any()).Return(nil)

	err := fetchFeeds(db, parser)
	assert.NoError(t, err)
}

func TestFetchFeedsOmitsItemsAfterItemLimit(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockFeed := randFeed()
	mockFeed.FetchLimit = 1

	mockFeedItems := test.MockItems()

	parser := mock_parser.NewMockParser(ctrl)
	parser.EXPECT().ParseURL(mockFeed.URL).Return(mockFeedItems, nil)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Feeds().Return([]*feed.Feed{mockFeed}, nil)
	db.EXPECT().MatchingItem(gomock.Any()).Return(nil, nil)
	db.EXPECT().SaveItem(gomock.Any()).Return(nil)

	err := fetchFeeds(db, parser)
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
		URL:        fmt.Sprintf("test url %d", rand.Int()),
		Tags:       randTags(2),
		FetchLimit: 5,
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

func randFeed() *feed.Feed {
	return &feed.Feed{
		URL: fmt.Sprintf("URL %d", rand.Int()),
	}
}

func randFeeds(n int) []*feed.Feed {
	feeds := make([]*feed.Feed, n, n)
	for i := 0; i < len(feeds); i++ {
		feeds[i] = randFeed()
	}
	return feeds
}

func mockFeeds() []*feed.Feed {
	return randFeeds(2)
}

func mockError() error {
	return errors.New("mock error")
}
