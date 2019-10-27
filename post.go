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
	FindByID(id PostID) (Post, error)
	Store(post Post) error
	FindLatestPostsForUser(id ID) ([]*Post, error)
	FindLatestPostsForUserAndFriends(user *User) ([]*Post, error)
}

type PostID string

type Author struct {
	UserID ID `bson:"user_id"`
}

type Post struct {
	ID        PostID `bson:"_id"`
	Author    Author
	Body      string
	Timestamp time.Time
}

func NewPost(author Author, body string) (*Post, error) {
	if body == "" {
		return nil, ErrEmptyBody
	}

	return &Post{Author: author, Body: body, Timestamp: time.Now()}, nil
}
