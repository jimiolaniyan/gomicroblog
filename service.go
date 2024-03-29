package blog

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jimiolaniyan/gomicroblog/auth"

	"github.com/rs/xid"
)

type Service interface {
	CreateProfile(id, username, email string)
	CreatePost(id ID, body string) (PostID, error)        //messaging
	GetUserPosts(username string) ([]*Post, error)        //messaging
	GetProfile(username string) (Profile, error)          //profile
	UpdateLastSeen(id ID) error                           //profile
	EditProfile(id ID, req editProfileRequest) error      //profile
	CreateRelationshipFor(id ID, username string) error   //profile
	RemoveRelationshipFor(id ID, username string) error   //profile
	GetUserFriends(username string) ([]UserInfo, error)   //profile
	GetUserFollowers(username string) ([]UserInfo, error) //profile
	GetTimeline(id ID) ([]postResponse, error)            //messaging
}

type service struct {
	users Repository
	posts PostRepository
}

type createPostRequest struct {
	Body string
}

type createPostResponse struct {
	ID PostID `json:"id"`
}

type editProfileRequest struct {
	Username *string
	Bio      *string
}

type Relationships struct {
	Followers int `json:"followers"`
	Friends   int `json:"following"`
}

type Profile struct {
	ID            ID             `json:"id"`
	Username      string         `json:"username"`
	Avatar        string         `json:"avatar_url,omitempty"`
	Bio           string         `json:"bio"`
	Joined        time.Time      `json:"joined"`
	LastSeen      time.Time      `json:"last_seen"`
	Relationships Relationships  `json:"relationships"`
	Posts         []postResponse `json:"posts"`
}

type authorResponse struct {
	UserID   ID     `json:"user_id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type postResponse struct {
	ID        PostID         `json:"id"`
	Body      string         `json:"body"`
	Timestamp time.Time      `json:"timestamp"`
	Author    authorResponse `json:"author"`
}

type profileResponse struct {
	Profile *Profile `json:"profile,omitempty"`
	URL     string   `json:"url"`
	Err     error    `json:"err,omitempty"`
}

type UserInfo struct {
	ID       ID        `json:"id"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar_url"`
	Bio      string    `json:"bio"`
	Joined   time.Time `json:"joined"`
}

func NewService(users Repository, posts PostRepository) Service {
	return &service{users: users, posts: posts}
}

func (svc *service) CreateProfile(id string, username string, email string) {
	if !svc.userExists(username, email) {
		now := time.Now().UTC()
		user := User{ID: ID(id), Username: username, Email: email, CreatedAt: now, LastSeen: now}
		_ = svc.users.Store(&user)
	}
}

func (svc *service) userExists(username string, email string) bool {
	if u, _ := svc.users.FindByName(username); u != nil {
		return true
	}
	if u, _ := svc.users.FindByEmail(email); u != nil {
		return true
	}
	return false
}

func (svc *service) CreatePost(id ID, body string) (PostID, error) {
	if !IsValidID(string(id)) {
		return "", ErrInvalidID
	}

	user, err := svc.users.FindByID(id)
	if err != nil {
		return "", err
	}

	author := Author{UserID: user.ID}
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

func (svc *service) GetUserPosts(username string) ([]*Post, error) {
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

	return Profile{
		ID:       user.ID,
		Username: username,
		Avatar:   avatar(user.Email),
		Bio:      user.Bio,
		Joined:   user.CreatedAt,
		LastSeen: user.LastSeen,
		Relationships: Relationships{
			Followers: len(user.Followers),
			Friends:   len(user.Friends),
		},
		Posts: buildPostResponses(posts, user),
	}, nil
}

func (svc *service) EditProfile(id ID, req editProfileRequest) error {
	if !IsValidID(string(id)) {
		return ErrInvalidID
	}

	if req.Username == nil && req.Bio == nil {
		return nil
	}

	user, err := svc.users.FindByID(id)
	if err != nil {
		return ErrNotFound
	}

	if req.Username != nil && *req.Username != user.Username {
		if err := svc.updateUsername(req.Username, user); err != nil {
			return err
		}
	}

	if req.Bio != nil {
		if err := user.UpdateBio(*req.Bio); err != nil {
			return err
		}
	}

	if err := svc.users.Update(user); err != nil {
		return err
	}

	return nil
}

func (svc *service) UpdateLastSeen(id ID) error {
	if !IsValidID(string(id)) {
		return ErrNotFound
	}

	user, err := svc.users.FindByID(id)
	if err != nil {
		return ErrNotFound
	}

	user.LastSeen = time.Now().UTC()
	err = svc.users.Update(user)
	if err != nil {
		return fmt.Errorf("error updating last seen: %s", err.Error())
	}
	return nil
}

func (svc *service) CreateRelationshipFor(id ID, username string) error {
	u1, u2, err := svc.getU1U2(id, username)
	if err != nil {
		return err
	}

	if u1.Username == username {
		return ErrCantFollowSelf
	}

	if u1.IsFollowing(u2) {
		return ErrAlreadyFollowing
	}

	u1.Follow(u2)

	if err = svc.users.Update(u1); err != nil {
		return err
	}

	if err = svc.users.Update(u2); err != nil {
		return err
	}

	return nil
}

func (svc *service) RemoveRelationshipFor(id ID, username string) error {
	u1, u2, err := svc.getU1U2(id, username)
	if err != nil {
		return err
	}

	if u1.Username == username {
		return ErrCantUnFollowSelf
	}

	if !u1.IsFollowing(u2) {
		return ErrNotFollowing
	}

	u1.Unfollow(u2)

	if err = svc.users.Update(u1); err != nil {
		return err
	}

	if err = svc.users.Update(u2); err != nil {
		return err
	}

	return nil
}

func (svc *service) GetUserFriends(username string) ([]UserInfo, error) {
	user, err := svc.findUser(username)
	if err != nil {
		return nil, err
	}

	if len(user.Friends) < 1 {
		return []UserInfo{}, nil
	}

	friends, err := svc.users.FindByIDs(user.Friends)
	if err != nil {
		return nil, err
	}

	return buildUserInfosFromUsers(friends), nil
}

func (svc *service) GetUserFollowers(username string) ([]UserInfo, error) {
	user, err := svc.findUser(username)
	if err != nil {
		return nil, err
	}

	if len(user.Followers) < 1 {
		return []UserInfo{}, nil
	}

	followers, err := svc.users.FindByIDs(user.Followers)
	if err != nil {
		return nil, err
	}

	return buildUserInfosFromUsers(followers), nil
}

func (svc *service) GetTimeline(id ID) ([]postResponse, error) {
	if !IsValidID(string(id)) {
		return nil, ErrInvalidID
	}

	user, err := svc.users.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}

	posts, _ := svc.posts.FindLatestPostsForUserAndFriends(user)

	return buildPostResponses(posts, user), nil
}

// TODO refactor this to use get U1 and U2 separately
func (svc *service) getU1U2(id ID, username string) (u1 *User, u2 *User, err error) {
	if !IsValidID(string(id)) {
		return nil, nil, ErrInvalidID
	}

	if username == "" {
		return nil, nil, ErrInvalidUsername
	}

	u1, err = svc.users.FindByID(id)
	if err != nil {
		return nil, nil, ErrNotFound
	}

	u2, err = svc.users.FindByName(username)
	if err != nil {
		return nil, nil, ErrNotFound
	}

	return
}

func (svc *service) findUser(username string) (*User, error) {
	if username == "" {
		return nil, ErrInvalidUsername
	}
	user, err := svc.users.FindByName(username)
	if err != nil {
		return nil, ErrNotFound
	}
	return user, nil
}

func buildUserInfosFromUsers(users []User) []UserInfo {
	var infos = []UserInfo{}
	for _, user := range users {
		infos = append(infos, UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Avatar:   avatar(user.Email),
			Bio:      user.Bio,
			Joined:   user.CreatedAt,
		})
	}
	return infos
}

func (svc *service) updateUsername(username *string, user *User) error {
	u := strings.TrimSpace(*username)
	if u == "" {
		return ErrInvalidUsername
	}
	if user, _ := svc.users.FindByName(u); user != nil {
		return ErrExistingUsername
	}
	user.Username = u
	return nil
}

func buildPostResponses(posts []*Post, user *User) []postResponse {
	var res = []postResponse{}

	for _, p := range posts {
		pr := postResponse{
			ID:        p.ID,
			Body:      p.Body,
			Timestamp: p.Timestamp,
			Author: authorResponse{
				UserID:   user.ID,
				Username: user.Username,
				Avatar:   avatar(user.Email),
			},
		}

		res = append(res, pr)
	}

	return res
}

func avatar(email string) string {
	digest := fmt.Sprintf("%x", md5.Sum([]byte(email)))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?d=identicon", digest)
}

type accountCreatedHandler struct {
	UserService Service
}

func (a accountCreatedHandler) AccountCreated(id string, username string, email string) {
	a.UserService.CreateProfile(id, username, email)
}

func NewAccountCreatedHandler(s Service) auth.Events {
	return accountCreatedHandler{UserService: s}
}
