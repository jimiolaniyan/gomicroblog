package auth

import (
	"errors"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/rs/xid"
)

type Account struct {
	ID          ID
	Credentials Credentials
	CreatedAt   time.Time
}

type ID string

//Credentials holds the account's sensitive information
type Credentials struct {
	Username,
	Email,
	Password string
}

var (
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrExistingUsername   = errors.New("username in use")
	ErrExistingEmail      = errors.New("email in use")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrNotFound           = errors.New("account not found")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

//NewAccount validates username and email and returns a new Account if
// arguments are valid
func NewAccount(username string, email string) (*Account, error) {
	r := regexp.MustCompile(`^\w{1,24}$`)
	if !r.MatchString(username) {
		return nil, ErrInvalidUsername
	}

	r = regexp.MustCompile(`^\S+@\S+\.\S+$`)
	if !r.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	c := Credentials{Username: username, Email: email}
	return &Account{Credentials: c}, nil
}

func NewID() ID {
	return ID(xid.New().String())
}

func isValidID(id string) bool {
	if _, err := xid.FromString(id); err != nil {
		return false
	}
	return true
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", errors.New("error hashing password")
	}
	return string(hash), nil
}

func hashMatchesPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
