package gomicroblog

import (
	"errors"
	"strings"
)

type Repository interface {
	FindByName(username string) (*user, error)
	Store(u *user) error
}

type user struct {
	username string
	password string
	email    string
}

var ErrNotFound = errors.New("user not found")

func NewUser(username, password, email string) (*user, error) {
	if err := validateArgs(username, password, email); err != nil {
		return nil, err
	}

	return &user{username: username, password: password, email: email}, nil
}

var ErrEmptyUserName = errors.New("username cannot be empty")
var ErrInvalidPassword = errors.New("invalid password")
var ErrInvalidEmail = errors.New("invalid email address")

func validateArgs(username string, password string, email string) error {
	if len(username) < 1 {
		return ErrEmptyUserName
	}
	if len(password) < 8 {
		return ErrInvalidPassword
	}
	if !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}

	return nil
}
