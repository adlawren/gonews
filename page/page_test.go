package page

import (
	"errors"
	"fmt"
	"gonews/feed"
	"gonews/mock_db"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewReturnsErrorWhenUnableToGetFeeds(t *testing.T) {
	ctrl := gomock.NewController(t)

	title := "Test Title"
	mockErr := mockError()

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Feeds().Return(nil, mockErr)

	pg, err := New(db, title, "")
	assert.Equal(t, pg.Title, title)
	expectedErrMsg := fmt.Sprintf(
		"failed to get feeds: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestNewReturnsErrorWhenUnableToGetItems(t *testing.T) {
	ctrl := gomock.NewController(t)

	title := "Test Title"
	mockErr := mockError()
	mockFeeds := randFeeds(2)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Feeds().Return(mockFeeds, nil)
	db.EXPECT().ItemsFromFeed(gomock.Any()).Return(nil, mockErr)

	pg, err := New(db, title, "")
	assert.Equal(t, pg.Title, title)
	expectedErrMsg := fmt.Sprintf(
		"failed to get items from feed: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestNewReturnsErrorWhenMatchingTagFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	title := "Test Title"
	mockErr := mockError()

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().MatchingTag(gomock.Any()).Return(nil, mockErr)

	pg, err := New(db, title, "tag 1")
	assert.Equal(t, pg.Title, title)
	expectedErrMsg := fmt.Sprintf(
		"failed to get tag: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestNewReturnsErrorWhenUnableToGetFeedsFromTag(t *testing.T) {
	ctrl := gomock.NewController(t)

	title := "Test Title"
	mockTag := randTag()
	mockErr := mockError()

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().MatchingTag(gomock.Any()).Return(mockTag, nil)
	db.EXPECT().FeedsFromTag(gomock.Any()).Return(nil, mockErr)

	pg, err := New(db, title, "tag 1")
	assert.Equal(t, pg.Title, title)
	expectedErrMsg := fmt.Sprintf(
		"failed to get feeds: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func TestNewReturnsPage(t *testing.T) {
	ctrl := gomock.NewController(t)

	title := "Test Title"
	mockFeeds := randFeeds(2)
	mockItems1 := randItems(2)
	mockItems2 := randItems(2)

	var mockItems []*feed.Item
	mockItems = append(mockItems, mockItems1...)
	mockItems = append(mockItems, mockItems2...)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Feeds().Return(mockFeeds, nil)
	db.EXPECT().ItemsFromFeed(gomock.Any()).Return(mockItems1, nil)
	db.EXPECT().ItemsFromFeed(gomock.Any()).Return(mockItems2, nil)

	pg, err := New(db, title, "")
	assert.Equal(t, pg.Title, title)
	assert.ElementsMatch(t, pg.Items, mockItems)
	assert.NoError(t, err)
}

func randTag() *feed.Tag {
	return &feed.Tag{
		Name: fmt.Sprintf("name %d", rand.Int()),
	}
}

func randFeeds(n int) []*feed.Feed {
	feeds := make([]*feed.Feed, n, n)
	for i := 0; i < len(feeds); i++ {
		feed := &feed.Feed{
			URL: fmt.Sprintf("test url %d", rand.Int()),
		}
		feeds[i] = feed
	}

	return feeds
}

func randItems(n int) []*feed.Item {
	now := time.Now()

	items := make([]*feed.Item, n, n)
	for i := 0; i < len(items); i++ {
		feedID := uint(rand.Int())
		item := &feed.Item{
			Title:       fmt.Sprintf("title %d", rand.Int()),
			Description: fmt.Sprintf("desc %d", rand.Int()),
			Link:        fmt.Sprintf("link %d", rand.Int()),
			Published:   now,
			Hide:        true,
			FeedID:      feedID,
		}
		item.Name = fmt.Sprintf("name %d", rand.Int())
		item.Email = fmt.Sprintf("email %d", rand.Int())

		items[i] = item
	}

	return items
}

func mockError() error {
	return errors.New("mock error")
}
