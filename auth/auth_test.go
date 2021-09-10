package auth_test // 'auth_test' instead of 'auth' to prevent gonews/test <- gonews/auth <- gonews/test import cycle

import (
	"fmt"
	"gonews/auth"
	"gonews/db/orm/query"
	"gonews/mock_db"
	"gonews/test"
	"gonews/user"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var (
	mockErr      = fmt.Errorf("mock error")
	mockUsername = test.MockUsername()
	mockPassword = test.MockPassword()
)

func TestIsValidReturnsErrorWhenDatabaseReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Find(gomock.Any(), gomock.Any()).DoAndReturn(func(ptr interface{}, clauses ...*query.Clause) error {
		_, ok := ptr.(*user.User)
		assert.True(t, ok)

		return mockErr
	})

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
	db.EXPECT().Find(gomock.Any(), gomock.Any()).DoAndReturn(func(ptr interface{}, clauses ...*query.Clause) error {
		_, ok := ptr.(*user.User)
		assert.True(t, ok)

		return query.ErrModelNotFound
	})

	isValid, err := auth.IsValid(mockUsername, mockPassword, db)
	assert.NoError(t, err)
	assert.False(t, isValid)
}

func TestIsValidReturnsFalseWhenHashDoesNotMatch(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockPasswordHash, err := auth.Hash("different_password")
	assert.NoError(t, err)

	mockUser := &user.User{
		Username:     mockUsername,
		PasswordHash: mockPasswordHash,
	}

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Find(gomock.Any(), gomock.Any()).DoAndReturn(func(ptr interface{}, clauses ...*query.Clause) error {
		user, ok := ptr.(*user.User)
		assert.True(t, ok)

		*user = *mockUser

		return nil
	})

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

	mockUser := &user.User{
		Username:     mockUsername,
		PasswordHash: mockPasswordHash,
	}

	db := mock_db.NewMockDB(ctrl)
	db.EXPECT().Find(gomock.Any(), gomock.Any()).DoAndReturn(func(ptr interface{}, clauses ...*query.Clause) error {
		user, ok := ptr.(*user.User)
		assert.True(t, ok)

		*user = *mockUser

		return nil
	})

	isValid, err := auth.IsValid(mockUsername, mockPassword, db)
	assert.NoError(t, err)
	assert.True(t, isValid)
}
