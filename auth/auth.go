package auth

import (
	"gonews/db"
	"gonews/user"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func IsValid(username, password string, db db.DB) (bool, error) {
	user, err := db.MatchingUser(&user.User{Username: username})
	if err != nil {
		return false, errors.Wrap(err, "failed to get matching user")
	}
	if user == nil {
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, errors.Wrap(err, "password invalid")
	}

	return err == nil, errors.Wrap(err, "failed to compare hash and password")
}

func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), errors.Wrap(err, "failed to generate password hash")
}
