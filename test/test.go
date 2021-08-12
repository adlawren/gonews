package test

import (
	"fmt"
	"gonews/auth"
	"gonews/config"
	"gonews/db"
	"gonews/feed"
	"gonews/user"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func tmpDB(t *testing.T) string {
	return fmt.Sprintf(
		"/tmp/gonews/test/%d/db.sqlite3",
		time.Now().UnixNano())
}

func InitDB(t *testing.T, migrationsDir string) (*config.DBConfig, db.DB) {
	tmpDB := tmpDB(t)
	log.Info().Msgf("Initializing test DB: %s", tmpDB)

	_, err := os.Stat(tmpDB)
	if os.IsNotExist(err) {
		goto createDB
	}

	assert.NoError(t, err)

createDB:
	err = os.MkdirAll(filepath.Dir(tmpDB), os.ModeDir)
	assert.NoError(t, err)

	_, err = os.Create(tmpDB)
	assert.NoError(t, err)

	dbCfg := &config.DBConfig{
		DSN: fmt.Sprintf("file:%s", tmpDB),
	}
	adb, err := db.New(dbCfg)
	assert.NoError(t, err)

	err = adb.Migrate(migrationsDir)
	assert.NoError(t, err)

	return dbCfg, adb
}

func MockUsername() string {
	return "mock_username"
}

func MockPassword() string {
	return "mock_password"
}

func mockPasswordHash(t *testing.T) string {
	mockPasswordHash, err := auth.Hash(MockPassword())
	assert.NoError(t, err)

	return mockPasswordHash
}

func MockUser(t *testing.T) *user.User {
	return &user.User{
		Username:     MockUsername(),
		PasswordHash: mockPasswordHash(t),
	}
}

func MockItem() *feed.Item {
	now := time.Now()
	return &feed.Item{
		Title:       fmt.Sprintf("Title %d", rand.Int()),
		Description: fmt.Sprintf("Description %d", rand.Int()),
		Link:        fmt.Sprintf("Link %d", rand.Int()),
		Published:   now,
		Name:        fmt.Sprintf("Name %d", rand.Int()),
		Email:       fmt.Sprintf("Email %d", rand.Int()),
	}
}

func MockItems() []*feed.Item {
	items := make([]*feed.Item, 2, 2)
	for i := 0; i < len(items); i++ {
		items[i] = MockItem()
	}
	return items
}

func MockFeed() *feed.Feed {
	return &feed.Feed{
		URL:             fmt.Sprintf("https://duckduckgo.com?q=%d", rand.Int()),
		FetchLimit:      uint(rand.Int()),
		AutoDismissedAt: time.Now(),
	}
}
