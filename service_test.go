package gomicroblog

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	svc    service
	req    registerUserRequest
	userID ID
	user   *user
}

func (ts *ServiceTestSuite) TearDownTest() {
	ts.svc.posts = NewPostRepository()
}

func (ts *ServiceTestSuite) SetupSuite() {
	ts.svc = service{users: NewUserRepository(), posts: NewPostRepository()}
	ts.req = registerUserRequest{"username", "password", "a@b"}

	id, _ := ts.svc.RegisterNewUser(ts.req)
	ts.userID = id

	user, _ := ts.svc.users.FindByID(id)
	ts.user = user
}

func (ts *ServiceTestSuite) TestService_RegisterNewUser() {
	now := time.Now().UTC()

	tests := []struct {
		description                string
		req                        *registerUserRequest
		wantValidID, wantCreatedAt bool
		wantLastSeen               bool
		wantErr                    error
	}{
		{
			description: "ExistingUsername",
			req:         &registerUserRequest{"username", "password1", "b@c"},
			wantErr:     ErrExistingUsername,
		},
		{
			description: "ExistingEmail",
			req:         &registerUserRequest{"username2", "password1", "a@b"},
			wantErr:     ErrExistingEmail,
		},
		{
			description: "InvalidPassword",
			req:         &registerUserRequest{"username2", "passwod", "b@c"},
			wantErr:     ErrInvalidPassword,
		},
		{
			description:   "ValidCredentials",
			req:           &registerUserRequest{"username2", "password", "b@c.com"},
			wantValidID:   true,
			wantCreatedAt: true,
			wantLastSeen:  true,
			wantErr:       nil,
		},
	}

	for _, tt := range tests {
		ts.Run(fmt.Sprintf("%s", tt.description), func() {
			userID, err := ts.svc.RegisterNewUser(*tt.req)

			assert.Equal(ts.T(), tt.wantErr, err)
			assert.Equal(ts.T(), IsValidID(string(userID)), tt.wantValidID)

			user, _ := ts.svc.users.FindByID(userID)
			if user != nil {
				assert.Equal(ts.T(), tt.wantCreatedAt, user.createdAt.After(now))
				assert.Equal(ts.T(), tt.wantLastSeen, user.lastSeen.After(now))
				assert.True(ts.T(), checkPasswordHash(user.password, "password"))
			}
		})
	}
}

func (ts ServiceTestSuite) TestService_ValidateUser() {

	tests := []struct {
		username, password string
		wantErr            error
		wantValidID        bool
	}{
		{"", "", ErrInvalidCredentials, false},
		{"user", "jaiu", ErrInvalidCredentials, false},
		{"nonexistent", "password", ErrInvalidCredentials, false},
		{"username", "incorrect", ErrInvalidCredentials, false},
		{"username", "password", nil, true},
	}

	for _, tt := range tests {
		req := validateUserRequest{tt.username, tt.password}

		userID, err := ts.svc.ValidateUser(req)

		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantValidID, IsValidID(string(userID)))
	}
}

func (ts *ServiceTestSuite) TestService_CreatePost() {
	tests := []struct {
		userID              ID
		body                string
		wantErr             error
		wantValidID, wantTS bool
		wantUsername        string
	}{
		{wantErr: ErrInvalidID},
		{userID: "user", wantErr: ErrInvalidID},
		{userID: nextID(), body: "post", wantErr: ErrNotFound},
		{userID: ts.userID, wantErr: ErrEmptyBody},
		{userID: ts.userID, body: "post", wantValidID: true, wantErr: nil, wantTS: true},
	}
	for _, tt := range tests {
		now := time.Now()
		id, err := ts.svc.CreatePost(tt.userID, tt.body)
		assert.Equal(ts.T(), tt.wantValidID, IsValidID(string(id)))
		assert.Equal(ts.T(), tt.wantErr, err)

		if tt.wantValidID {
			post, err := ts.svc.posts.FindByID(id)
			assert.Nil(ts.T(), err)
			assert.Equal(ts.T(), tt.body, post.body)
			assert.Equal(ts.T(), tt.userID, post.Author.UserID)
			assert.Equal(ts.T(), tt.wantTS, post.timestamp.After(now))
		}
	}
}

func (ts *ServiceTestSuite) TestService_GetUserPosts() {
	_, _ = ts.svc.CreatePost(ts.userID, "body")

	tests := []struct {
		username     string
		wantErr      error
		wantPostsLen int
	}{
		{wantErr: ErrInvalidUsername},
		{username: "void", wantErr: ErrNotFound},
		{username: "username", wantErr: nil, wantPostsLen: 1},
	}

	for _, tt := range tests {
		posts, err := ts.svc.GetUserPosts(tt.username)
		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantPostsLen, len(posts))
	}
}

func (ts *ServiceTestSuite) TestService_GetProfile() {

	av := avatar(ts.req.Email)
	username := ts.req.Username
	tests := []struct {
		username                 string
		wantErr                  error
		wantUsername, wantAvatar string
		wantID                   ID
		wantJoined, wantLastSeen bool
	}{
		{username: "", wantErr: ErrInvalidUsername, wantUsername: ""},
		{username: "void", wantErr: ErrNotFound, wantUsername: ""},
		{username: username, wantErr: nil, wantUsername: username, wantAvatar: av, wantJoined: true, wantLastSeen: true, wantID: ts.user.ID},
	}

	for _, tt := range tests {
		p, err := ts.svc.GetProfile(tt.username)

		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantUsername, p.Username)
		assert.Equal(ts.T(), tt.wantAvatar, p.Avatar)
		assert.Equal(ts.T(), tt.wantID, p.ID)

		if tt.wantErr == nil {
			assert.Equal(ts.T(), ts.user.createdAt, p.Joined)
			assert.Equal(ts.T(), ts.user.lastSeen, p.LastSeen)
		}
	}
}

func (ts *ServiceTestSuite) TestService_UpdateLastSeen() {
	tests := []struct {
		userID  ID
		wantErr error
		wantLS  bool
	}{
		{wantErr: ErrNotFound},
		{userID: nextID(), wantErr: ErrNotFound},
		{userID: ts.userID, wantLS: true},
	}
	now := time.Now().UTC()
	for _, tt := range tests {
		err := ts.svc.UpdateLastSeen(tt.userID)
		assert.Equal(ts.T(), tt.wantErr, err)

		if tt.wantLS {
			assert.Equal(ts.T(), tt.wantLS, ts.user.lastSeen.After(now))
		}
	}
}

func (ts *ServiceTestSuite) TestEditProfile() {
	r := editProfileRequest{Username: "U", Bio: "My new bio"}
	longBio := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut " +
		"labore et dolore magna aliqua. Ut enim ad minim h"

	newUser := *ts.user
	newUser.ID = nextID()
	newUser.username = "newUsername"

	err := ts.svc.users.Store(&newUser)
	assert.Nil(ts.T(), err)

	tests := []struct {
		id      ID
		req     editProfileRequest
		wantErr error
	}{
		{wantErr: ErrInvalidID},
		{id: nextID(), wantErr: ErrInvalidUsername},
		{id: nextID(), req: r, wantErr: ErrNotFound},
		{id: newUser.ID, req: editProfileRequest{"U", longBio}, wantErr: ErrBioTooLong},
		{id: newUser.ID, req: r, wantErr: nil},
	}

	for _, tt := range tests {
		err := ts.svc.EditProfile(tt.id, tt.req)
		assert.Equal(ts.T(), tt.wantErr, err)

		if tt.wantErr == nil {
			assert.Equal(ts.T(), tt.req.Username, newUser.username)
			assert.Equal(ts.T(), tt.req.Bio, newUser.bio)
		}
	}
}

func (ts *ServiceTestSuite) TestNewService() {
	users := NewUserRepository()
	posts := NewPostRepository()
	svc := NewService(users, posts)

	s := svc.(*service)

	assert.Equal(ts.T(), users, s.users)
	assert.Equal(ts.T(), posts, s.posts)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
