package db_test // 'db_test' instead of 'db' to prevent gonews/test <- gonews/db <- gonews/test import cycle

import (
	"fmt"
	"gonews/feed"
	"gonews/test"
	"gonews/user"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	migrationsDir = "./migrations"
)

func TestMatchingItemReturnsMatchingItem(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockItem := test.MockItem()
	err := testDB.SaveItem(mockItem)
	assert.NoError(t, err)

	item, err := testDB.MatchingItem(mockItem)
	assert.NoError(t, err)

	assert.NotEqual(t, 0, item.ID)
	assertItemsEqual(t, mockItem, item)
}

func TestMatchingItemReturnsNilWhenNoMatchingItemExists(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockItem := test.MockItem()

	item, err := testDB.MatchingItem(mockItem)
	assert.NoError(t, err)
	assert.Nil(t, item)
}

func TestSaveItemInsertsNewItem(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockItem := test.MockItem()
	err := testDB.SaveItem(mockItem)
	assert.NoError(t, err)

	items, err := testDB.Items()
	assert.NoError(t, err)

	assert.Len(t, items, 1)

	item := items[0]
	assert.NotEqual(t, 0, item.ID)
	assertItemsEqual(t, mockItem, item)
}

func TestSaveItemUpdatesExistingItem(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockItem := test.MockItem()
	err := testDB.SaveItem(mockItem)
	assert.NoError(t, err)

	mockItem.Title = fmt.Sprintf("Updated %s", mockItem.Title)
	err = testDB.SaveItem(mockItem)
	assert.NoError(t, err)

	items, err := testDB.Items()
	assert.NoError(t, err)

	assert.Len(t, items, 1)

	item := items[0]
	assert.NotEqual(t, 0, item.ID)
	assertItemsEqual(t, mockItem, item)
}

func TestSaveItemSetsCreatedAtToCurrentTimeForNewItem(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockItem := test.MockItem()
	currentTime := time.Now()
	err := testDB.SaveItem(mockItem)
	assert.NoError(t, err)

	items, err := testDB.Items()
	assert.NoError(t, err)

	assert.Len(t, items, 1)

	item := items[0]
	assert.True(t, currentTime.Before(item.CreatedAt))
}

func TestSaveItemDoesNotChangeCreatedAtDuringUpdate(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockItem := test.MockItem()
	err := testDB.SaveItem(mockItem)
	assert.NoError(t, err)

	items, err := testDB.Items()
	assert.NoError(t, err)

	assert.Len(t, items, 1)

	item := items[0]
	createdAt := item.CreatedAt

	err = testDB.SaveItem(mockItem)
	assert.NoError(t, err)
	items, err = testDB.Items()
	assert.NoError(t, err)

	assert.Len(t, items, 1)

	itemCopy := items[0]
	assert.Equal(t, createdAt, itemCopy.CreatedAt)
}

func TestFeeds(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockFeed := test.MockFeed()
	err := testDB.SaveFeed(mockFeed)
	assert.NoError(t, err)

	feeds, err := testDB.Feeds()
	assert.NoError(t, err)

	assert.Len(t, feeds, 1)

	feed := feeds[0]
	assert.NotEqual(t, 0, feed.ID)
	assertFeedsEqual(t, mockFeed, feed)
}

func TestSaveFeedInsertsNewFeed(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockFeed := test.MockFeed()
	err := testDB.SaveFeed(mockFeed)
	assert.NoError(t, err)

	feeds, err := testDB.Feeds()
	assert.NoError(t, err)

	assert.Len(t, feeds, 1)

	feed := feeds[0]
	assert.NotEqual(t, 0, feed.ID)
	assertFeedsEqual(t, mockFeed, feed)
}

func TestSaveFeedUpdatesExistingFeed(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockFeed := test.MockFeed()
	err := testDB.SaveFeed(mockFeed)
	assert.NoError(t, err)

	mockFeed.URL = fmt.Sprintf("%s&updated=yes", mockFeed.URL)
	err = testDB.SaveFeed(mockFeed)
	assert.NoError(t, err)

	feeds, err := testDB.Feeds()
	assert.NoError(t, err)

	assert.Len(t, feeds, 1)

	feed := feeds[0]
	assert.NotEqual(t, 0, feed.ID)
	assertFeedsEqual(t, mockFeed, feed)
}

func assertUsersEqual(t *testing.T, u1, u2 *user.User) {
	assert.Equal(t, u1.Username, u2.Username)
	assert.Equal(t, u1.PasswordHash, u2.PasswordHash)
}

func assertItemsEqual(t *testing.T, i1, i2 *feed.Item) {
	assert.Equal(t, i1.Name, i2.Name)
	assert.Equal(t, i1.Email, i2.Email)
	assert.Equal(t, i1.Title, i2.Title)
	assert.Equal(t, i1.Description, i2.Description)
	assert.Equal(t, i1.Link, i2.Link)
	assert.Equal(t, i1.Published.UnixNano(), i2.Published.UnixNano())
	assert.Equal(t, i1.Hide, i2.Hide)
	assert.Equal(t, i1.FeedID, i2.FeedID)
}

func assertFeedsEqual(t *testing.T, f1, f2 *feed.Feed) {
	assert.Equal(t, f1.URL, f2.URL)
	assert.Equal(t, f1.FetchLimit, f2.FetchLimit)
}
