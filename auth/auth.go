package auth

import (
	"errors"
	"fmt"
	"gonews/db"
	"gonews/db/orm/query"
	"gonews/db/orm/query/clause"
	"gonews/user"

	"golang.org/x/crypto/bcrypt"
)

func IsValid(username, password string, db db.DB) (bool, error) {
	var user user.User
	err := db.Find(&user, clause.New("where username = ?", username))
	if errors.Is(err, query.ErrModelNotFound) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to get matching user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, fmt.Errorf("password invalid: %w", err)
	}
	if err != nil {
		return false, fmt.Errorf("failed to compare hash and password: %w", err)
	}

	return true, nil
}

func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}

	return string(hash), nil
}
