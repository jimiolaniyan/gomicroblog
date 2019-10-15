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
	FindPostsByName(username string) ([]*post, error)
}

type PostID string

type Author struct {
	UserID   ID
	Username string
	Avatar   string
}

type post struct {
	ID        PostID
	Author    Author
	body      string
	timestamp time.Time
}

// TODO check that author id and name are properly set
func NewPost(author Author, body string) (*post, error) {
	if body == "" {
		return nil, ErrEmptyBody
	}

	return &post{Author: author, body: body, timestamp: time.Now()}, nil
}
