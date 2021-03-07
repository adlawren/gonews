package test

import (
	"fmt"
	"gonews/auth"
	"gonews/config"
	"gonews/db"
	"gonews/user"
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
