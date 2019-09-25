package gomicroblog

import (
	"errors"
	"strings"
)

type Repository interface {
	FindByName(username string) (*user, error)
	Store(u *user) error
	FindByEmail(e string) (*user, error)
}

type ID string

// TODO: consider moving username, password
//  and email to a credentials value object
type user struct {
	ID       ID
	username string
	password string
	email    string
}

//type Credentials struct{
//	username string
//	password string
//	email    string
//}

var (
	ErrNotFound        = errors.New("user not found")
	ErrEmptyUserName   = errors.New("username cannot be empty")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidEmail    = errors.New("invalid email address")
)

func NewUser(username, password, email string) (*user, error) {
	if err := validateArgs(username, password, email); err != nil {
		return nil, err
	}

	return &user{username: username, password: password, email: email}, nil
}

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
