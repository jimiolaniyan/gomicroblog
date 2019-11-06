package blog

import (
	"errors"
	"strings"
	"time"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	FindByName(username string) (*User, error)
	Store(u *User) error
	FindByEmail(e string) (*User, error)
	FindByID(id ID) (*User, error)
	Delete(id ID) error
	Update(u *User) error
	FindByIDs(ids []ID) ([]User, error)
}

type ID string

type User struct {
	ID        ID `bson:"_id"`
	Username  string
	Password  string
	Email     string
	CreatedAt time.Time
	LastSeen  time.Time
	Bio       string
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

func (u *User) IsFollowing(u2 *User) bool {
	for _, id := range u.Friends {
		if id == u2.ID {
			return true
		}
	}

	return false
}

func (u *User) Follow(u2 *User) {
	u.Friends = append(u.Friends, u2.ID)
	u2.Followers = append(u2.Followers, u.ID)
}

func (u *User) Unfollow(u2 *User) {
	// remove u2 from u1 friends
	for i, id := range u.Friends {
		if id == u2.ID {
			u.Friends = append(u.Friends[:i], u.Friends[i+1:]...)
			break
		}
	}

	// remove u1 from u2 followers
	for i, id := range u2.Followers {
		if id == u.ID {
			u2.Followers = append(u2.Followers[:i], u2.Followers[i+1:]...)
			break
		}
	}
}

func (u *User) UpdateBio(bio string) error {
	b := strings.TrimSpace(bio)
	if len(b) > 140 {
		return ErrBioTooLong
	}
	u.Bio = b
	return nil
}

func NewUser(username, email string) (*User, error) {
	if err := validateArgs(username, email); err != nil {
		return nil, err
	}

	return &User{Username: username, Email: email}, nil
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
