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
	Update(u *user) error
	FindByIDs(ids []ID) ([]user, error)
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
	Friends   []ID
	Followers []ID
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
	ErrCantUnFollowSelf   = errors.New("can't unfollow yourself")
	ErrAlreadyFollowing   = errors.New("already following user")
	ErrNotFollowing       = errors.New("not following user")
)

func (u1 *user) IsFollowing(u2 *user) bool {
	for _, id := range u1.Friends {
		if id == u2.ID {
			return true
		}
	}

	return false
}

func (u1 *user) Follow(u2 *user) {
	u1.Friends = append(u1.Friends, u2.ID)
	u2.Followers = append(u2.Followers, u1.ID)
}

func (u1 *user) Unfollow(u2 *user) {
	// remove u2 from u1 friends
	for i, id := range u1.Friends {
		if id == u2.ID {
			u1.Friends = append(u1.Friends[:i], u1.Friends[i+1:]...)
			break
		}
	}

	// remove u1 from u2 followers
	for i, id := range u2.Followers {
		if id == u1.ID {
			u2.Followers = append(u2.Followers[:i], u2.Followers[i+1:]...)
			break
		}
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

//IsValidID checks if a given id is valid based on the xid library definition of a valid id
// this method should change if we ever change our uid generation library
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
