package auth_test // 'auth_test' instead of 'auth' to prevent gonews/test <- gonews/auth <- gonews/test import cycle

import (
	"errors"
	"fmt"
	"gonews/auth"
	"gonews/mock_db"
	"gonews/test"
	"gonews/user"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var (
	mockErr      = errors.New("mock error")
	mockUsername = test.MockUsername()
	mockPassword = test.MockPassword()
)

func TestIsValidReturnsErrorWhenDatabaseReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().MatchingUser(gomock.Any()).Return(nil, mockErr)

	isValid, err := auth.IsValid(mockUsername, mockPassword, db)
	expectedErrMsg := fmt.Sprintf(
		"failed to get matching user: %v",
		mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
	assert.False(t, isValid)
}

func TestIsValidReturnsFalseWhenUserDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().MatchingUser(gomock.Any()).Return(nil, nil)

	isValid, err := auth.IsValid(mockUsername, mockPassword, db)
	assert.NoError(t, err)
	assert.False(t, isValid)
}

func TestIsValidReturnsFalseWhenHashDoesNotMatch(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockPasswordHash, err := auth.Hash("different_password")
	assert.NoError(t, err)

	mock_user := &user.User{
		Username:     mockUsername,
		PasswordHash: mockPasswordHash,
	}

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().MatchingUser(gomock.Any()).Return(mock_user, nil)

	isValid, err := auth.IsValid(mockUsername, mockPassword, db)
	expectedErrMsg := fmt.Sprintf(
		"password invalid: %v",
		bcrypt.ErrMismatchedHashAndPassword)
	assert.EqualError(t, err, expectedErrMsg)
	assert.False(t, isValid)
}

func TestIsValidReturnsTrueWhenHashMatches(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockPasswordHash, err := auth.Hash(mockPassword)
	assert.NoError(t, err)

	mock_user := &user.User{
		Username:     mockUsername,
		PasswordHash: mockPasswordHash,
	}

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().MatchingUser(gomock.Any()).Return(mock_user, nil)

	isValid, err := auth.IsValid(mockUsername, mockPassword, db)
	assert.NoError(t, err)
	assert.True(t, isValid)
}
