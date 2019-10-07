package gomicroblog

import (
	"errors"
	"time"
)

var (
	ErrEmptyBody    = errors.New("post body cannot be empty")
	ErrPostNotFound = errors.New("post not found")
)

type PostRepository interface {
	FindByID(id PostID) (*post, error)
	Store(post *post) error
	FindByUserID(id ID) ([]*post, error)
}

type PostID string

type post struct {
	ID        PostID
	UserID    ID
	body      string
	timestamp time.Time
}

func NewPost(userID ID, body string) (*post, error) {
	if body == "" {
		return nil, ErrEmptyBody
	}

	return &post{UserID: userID, body: body, timestamp: time.Now()}, nil
}
