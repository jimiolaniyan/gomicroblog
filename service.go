package gomicroblog

import (
	"errors"
	"fmt"

	"github.com/rs/xid"
)

type Service interface {
	RegisterNewUser(req registerUserRequest) (ID, error)
	ValidateUser(req validateUserRequest) (ID, error)
	CreatePost(id ID, body string) (PostID, error)
	GetUserPosts(username string) ([]*post, error)
}

type service struct {
	users Repository
	posts PostRepository
}

type registerUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type validateUserRequest struct {
	Username, Password string
}

type createPostRequest struct {
	Body string
}

type registerUserResponse struct {
	ID  ID    `json:"id,omitempty"`
	Err error `json:"error,omitempty"`
}

type createPostResponse struct {
	ID PostID `json:"id"`
}

func (svc *service) RegisterNewUser(req registerUserRequest) (ID, error) {
	u := req.Username
	e := req.Email
	user, err := NewUser(u, e)
	if err != nil {
		return "", err
	}

	p := req.Password
	if len(p) < 8 {
		return "", ErrInvalidPassword
	}

	if _, err := verifyNotInUse(svc, u, e); err != nil {
		return "", err
	}

	user.ID = nextID()
	if hash, err := hashPassword(p); err == nil {
		user.password = hash
	}

	if err = svc.users.Store(user); err != nil {
		return "", fmt.Errorf("error saving user: %s ", err)
	}

	return user.ID, nil
}

func (svc *service) ValidateUser(req validateUserRequest) (ID, error) {
	if req.Username == "" || len(req.Password) < 8 {
		return "", ErrInvalidCredentials
	}

	user, err := svc.users.FindByName(req.Username)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !checkPasswordHash(user.password, req.Password) {
		return "", ErrInvalidCredentials
	}

	return user.ID, nil
}

func verifyNotInUse(svc *service, username string, email string) (*user, error) {
	if u, _ := svc.users.FindByName(username); u != nil {
		return nil, ErrExistingUsername
	}
	if u, _ := svc.users.FindByEmail(email); u != nil {
		return nil, ErrExistingEmail
	}
	return nil, nil
}

func (svc *service) CreatePost(id ID, body string) (PostID, error) {
	if !IsValidID(string(id)) {
		return "", ErrInvalidID
	}
	user, err := svc.users.FindByID(id)
	if err != nil {
		return "", err
	}

	author := Author{UserID: id, Username: user.username}
	post, err := NewPost(author, body)
	if err != nil {
		return "", err
	}

	// TODO refactor this to return next id
	post.ID = PostID(xid.New().String())
	if err = svc.posts.Store(post); err != nil {
		return "", errors.New("error saving post")
	}

	return post.ID, nil
}

func (svc *service) GetUserPosts(username string) ([]*post, error) {
	if username == "" {
		return []*post{}, ErrInvalidUsername
	}

	return svc.posts.FindPostsByName(username)
}

func NewService(users Repository, posts PostRepository) Service {
	return &service{users: users, posts: posts}
}
