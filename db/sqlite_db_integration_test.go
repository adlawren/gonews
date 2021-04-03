package db_test // 'db_test' instead of 'db' to prevent gonews/test <- gonews/db <- gonews/test import cycle

import (
	"gonews/feed"
	"gonews/test"
	"gonews/user"
	"testing"

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

func TestUsers(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockUser := test.MockUser(t)
	err := testDB.SaveUser(mockUser)
	assert.NoError(t, err)

	users, err := testDB.Users()
	assert.NoError(t, err)

	assert.Len(t, users, 1)

	user := users[0]
	assert.NotEqual(t, 0, user.ID)
	assertUsersEqual(t, mockUser, user)
}

func TestMatchingUserReturnsMatchingUser(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockUser := test.MockUser(t)
	err := testDB.SaveUser(mockUser)
	assert.NoError(t, err)

	user, err := testDB.MatchingUser(mockUser)
	assert.NoError(t, err)

	assert.NotEqual(t, 0, user.ID)
	assertUsersEqual(t, mockUser, user)
}

func TestMatchingUserReturnsNilWhenNoMatchingUserExists(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockUser := test.MockUser(t)

	user, err := testDB.MatchingUser(mockUser)
	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestSaveUserUpdatesExistingUserWithTheSameID(t *testing.T) {
	_, testDB := test.InitDB(t, migrationsDir)

	mockUser := test.MockUser(t)
	err := testDB.SaveUser(mockUser)
	assert.NoError(t, err)

	mockUser.Username = "different_username"
	err = testDB.SaveUser(mockUser)
	assert.NoError(t, err)

	user, err := testDB.MatchingUser(mockUser)
	assert.NoError(t, err)

	assert.Equal(t, mockUser.ID, user.ID)
	assertUsersEqual(t, mockUser, user)
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
