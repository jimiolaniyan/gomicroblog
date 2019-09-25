package gomicroblog

import (
	"errors"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
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

func NewUser(username, email string) (*user, error) {
	if err := validateArgs(username, email); err != nil {
		return nil, err
	}

	return &user{username: username, email: email}, nil
}

func validateArgs(username string, email string) error {
	if len(username) < 1 {
		return ErrEmptyUserName
	}

	if !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}

	return nil
}

func nextID() ID {
	return ID(xid.New().String())
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", errors.New("error hashing password")
	}
	return string(hash), nil
}

func checkPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
