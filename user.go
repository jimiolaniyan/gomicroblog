package gomicroblog

import (
	"errors"
	"strings"
	"time"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	FindByName(username string) (*user, error)
	Store(u *user) error
	FindByEmail(e string) (*user, error)
	FindByID(id ID) (*user, error)
	Delete(id ID) error
}

type ID string

// TODO: consider moving username, password
//  and email to a credentials value object
//  as part of an auth service
type user struct {
	ID        ID
	username  string
	password  string
	email     string
	createdAt time.Time
	lastSeen  time.Time
	bio       string
	Friends   []*user
	Followers []*user
}

var (
	ErrInvalidID          = errors.New("invalid user id")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrNotFound           = errors.New("user not found")
	ErrExistingUsername   = errors.New("username in use")
	ErrExistingEmail      = errors.New("email in use")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrBioTooLong         = errors.New("bio cannot be more than 140 characters")
	ErrCantFollowSelf     = errors.New("can't follow yourself")
)

func (u1 *user) IsFollowing(u2 *user) bool {
	for _, friend := range u1.Friends {
		if friend.ID == u2.ID {
			return true
		}
	}
	return false
}

func (u1 *user) Follow(u2 *user) {
	if !u1.IsFollowing(u2) {
		u1.Friends = append(u1.Friends, u2)
		u2.Followers = append(u2.Followers, u1)
	}
}

func NewUser(username, email string) (*user, error) {
	if err := validateArgs(username, email); err != nil {
		return nil, err
	}

	return &user{username: username, email: email}, nil
}

func validateArgs(username string, email string) error {
	if len(username) < 1 {
		return ErrInvalidUsername
	}

	if !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}

	return nil
}

func nextID() ID {
	return ID(xid.New().String())
}

//IsValidID checks if a given id is valid based on the xid library
func IsValidID(id string) bool {
	if _, err := xid.FromString(id); err == xid.ErrInvalidID {
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

func checkPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
