package gomicroblog

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/xid"
)

type Service interface {
	RegisterNewUser(req registerUserRequest) (ID, error)
	ValidateUser(req validateUserRequest) (ID, error)
	CreatePost(id ID, body string) (PostID, error)
	GetUserPosts(username string) ([]*post, error)
	GetProfile(username string) (Profile, error)
	UpdateLastSeen(id ID) error
	EditProfile(id ID, req editProfileRequest) error
}

type service struct {
	users Repository
	posts PostRepository
}

type registerUserRequest struct {
	Username string
	Password string
	Email    string
}

type validateUserRequest struct {
	Username, Password string
}

type createPostRequest struct {
	Body string
}

type editProfileRequest struct {
	Username string
	Bio      string
}

type registerUserResponse struct {
	ID  ID    `json:"id,omitempty"`
	Err error `json:"error,omitempty"`
}

type createPostResponse struct {
	ID PostID `json:"id"`
}

type Relationships struct {
	Followers []user `json:"followers"`
	Following []user `json:"following"`
}

type Profile struct {
	ID            ID            `json:"id"`
	Username      string        `json:"username"`
	Avatar        string        `json:"avatar_url,omitempty"`
	Bio           string        `json:"bio"`
	Joined        time.Time     `json:"joined"`
	LastSeen      time.Time     `json:"last_seen"`
	Relationships Relationships `json:"relationships"`
	Posts         []*post       `json:"posts"`
}

type profileResponse struct {
	Profile *Profile `json:"profile,omitempty"`
	URL     string   `json:"url"`
	Err     error    `json:"err,omitempty"`
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

	now := time.Now().UTC()
	user.createdAt = now
	user.lastSeen = now
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
	_, err := svc.users.FindByID(id)
	if err != nil {
		return "", err
	}

	author := Author{UserID: id}
	post, err := NewPost(author, body)
	if err != nil {
		return "", err
	}

	// TODO refactor this to return next id
	post.ID = PostID(xid.New().String())
	if err = svc.posts.Store(*post); err != nil {
		return "", errors.New("error saving post")
	}

	return post.ID, nil
}

func (svc *service) GetUserPosts(username string) ([]*post, error) {
	if username == "" {
		return nil, ErrInvalidUsername
	}

	user, err := svc.users.FindByName(username)
	if err != nil {
		return nil, ErrNotFound
	}

	return svc.posts.FindLatestPostsForUser(user.ID)
}

func (svc *service) GetProfile(username string) (Profile, error) {
	if username == "" {
		return Profile{}, ErrInvalidUsername
	}

	user, err := svc.users.FindByName(username)
	if err != nil {
		return Profile{}, ErrNotFound
	}

	posts, err := svc.posts.FindLatestPostsForUser(user.ID)
	if err != nil {
		return Profile{}, errors.New("error finding latest posts")
	}

	if len(posts) > 0 {
		addAuthorInfoToPosts(posts, user)
	}

	return Profile{
		ID:       user.ID,
		Username: username,
		Avatar:   avatar(user.email),
		Joined:   user.createdAt,
		LastSeen: user.lastSeen,
		Posts:    posts,
	}, nil
}

func (svc *service) UpdateLastSeen(id ID) error {
	if !IsValidID(string(id)) {
		return ErrInvalidID
	}

	user, err := svc.users.FindByID(id)
	if err != nil {
		return ErrNotFound
	}

	user.lastSeen = time.Now().UTC()
	err = svc.users.Store(user)
	if err != nil {
		return fmt.Errorf("error updating last seen: %s", err.Error())
	}
	return nil
}

func addAuthorInfoToPosts(posts []*post, user *user) {
	for _, p := range posts {
		p.Author.Avatar = avatar(user.email)
		p.Author.Username = user.username
	}
}

func avatar(email string) string {
	digest := fmt.Sprintf("%x", md5.Sum([]byte(email)))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?d=identicon", digest)
}

func (svc *service) EditProfile(id ID, req editProfileRequest) error {
	if !IsValidID(string(id)) {
		return ErrInvalidID
	}

	if req.Username == "" {
		return ErrInvalidUsername
	}

	bio := strings.TrimSpace(req.Bio)
	if len(bio) > 140 {
		return ErrBioTooLong
	}

	user, err := svc.users.FindByID(id)
	if err != nil {
		return ErrNotFound
	}

	user.username = req.Username
	user.bio = bio

	return nil
}

func NewService(users Repository, posts PostRepository) Service {
	return &service{users: users, posts: posts}
}
