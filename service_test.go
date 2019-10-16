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
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.svc = service{users: NewUserRepository()}
	suite.req = registerUserRequest{"username", "password", "a@b"}
	id, _ := suite.svc.RegisterNewUser(suite.req)
	suite.userID = id
}

func TestService_RegisterNewUser(t *testing.T) {
	now := time.Now().UTC()
	svc := service{users: NewUserRepository()}
	req := registerUserRequest{"username", "password", "a@b"}

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
		t.Run(fmt.Sprintf("%s", tt.description), func(t *testing.T) {
			_, err := svc.RegisterNewUser(req)
			userID, err := svc.RegisterNewUser(*tt.req)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, IsValidID(string(userID)), tt.wantValidID)

			user, _ := svc.users.FindByID(userID)
			if user != nil {
				assert.Equal(t, tt.wantCreatedAt, user.createdAt.After(now))
				assert.Equal(t, tt.wantLastSeen, user.lastSeen.After(now))
			}
		})
	}
}

func TestService_ValidateUser(t *testing.T) {
	svc := service{users: NewUserRepository()}
	_, err := svc.RegisterNewUser(registerUserRequest{"user", "password", "a@b.com"})
	assert.Nil(t, err)

	tests := []struct {
		username, password string
		wantErr            error
		wantValidID        bool
	}{
		{"", "", ErrInvalidCredentials, false},
		{"user", "jaiu", ErrInvalidCredentials, false},
		{"nonexistent", "password", ErrInvalidCredentials, false},
		{"user", "incorrect", ErrInvalidCredentials, false},
		{"user", "password", nil, true},
	}

	for _, tt := range tests {
		req := validateUserRequest{tt.username, tt.password}

		userID, err := svc.ValidateUser(req)

		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantValidID, IsValidID(string(userID)))
	}
}

func TestService_CreatePost(t *testing.T) {
	svc := NewService(NewUserRepository(), NewPostRepository())
	id, _ := svc.RegisterNewUser(registerUserRequest{"user", "password", "e@mail.com"})
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
		{userID: id, wantErr: ErrEmptyBody},
		{userID: id, body: "post", wantValidID: true, wantErr: nil, wantTS: true},
	}
	for _, tt := range tests {
		ts := time.Now()
		id, err := svc.CreatePost(tt.userID, tt.body)
		assert.Equal(t, tt.wantValidID, IsValidID(string(id)))
		assert.Equal(t, tt.wantErr, err)

		if tt.wantValidID {
			post, err := svc.(*service).posts.FindByID(id)
			assert.Nil(t, err)
			assert.Equal(t, tt.body, post.body)
			assert.Equal(t, tt.userID, post.Author.UserID)
			assert.Equal(t, tt.wantTS, post.timestamp.After(ts))
		}
	}
}

func TestService_GetUserPosts(t *testing.T) {
	svc := NewService(NewUserRepository(), NewPostRepository())
	id, _ := svc.RegisterNewUser(registerUserRequest{"user", "password", "e@mail.com"})
	_, _ = svc.CreatePost(id, "body")

	tests := []struct {
		username     string
		wantErr      error
		wantPostsLen int
	}{
		{wantErr: ErrInvalidUsername},
		{username: "void", wantErr: ErrNotFound},
		{username: "user", wantErr: nil, wantPostsLen: 1},
	}

	for _, tt := range tests {
		posts, err := svc.GetUserPosts(tt.username)
		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantPostsLen, len(posts))
	}
}

func TestService_GetProfile(t *testing.T) {
	svc := NewService(NewUserRepository(), NewPostRepository())
	email := "e@mail.com"
	userID, _ := svc.RegisterNewUser(registerUserRequest{"user", "password", email})
	user, _ := svc.(*service).users.FindByID(userID)

	tests := []struct {
		username                 string
		wantErr                  error
		wantUsername, wantAvatar string
		wantJoined, wantLastSeen bool
	}{
		{username: "", wantErr: ErrInvalidUsername, wantUsername: ""},
		{username: "void", wantErr: ErrNotFound, wantUsername: ""},
		{username: "user", wantErr: nil, wantUsername: "user", wantAvatar: avatar(email), wantJoined: true, wantLastSeen: true},
	}

	for _, tt := range tests {
		p, err := svc.GetProfile(tt.username)

		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantUsername, p.Username)
		assert.Equal(t, tt.wantAvatar, p.Avatar)

		if tt.wantJoined || tt.wantLastSeen {
			assert.Equal(t, user.createdAt, p.Joined)
			assert.Equal(t, user.lastSeen, p.LastSeen)
		}
	}
}

func (suite *ServiceTestSuite) TestService_UpdateLastSeen() {
	tests := []struct {
		userID  ID
		wantErr error
		wantLS  bool
	}{
		{wantErr: ErrInvalidID},
		{userID: nextID(), wantErr: ErrNotFound},
		{userID: suite.userID, wantLS: true},
	}
	now := time.Now().UTC()
	for _, tt := range tests {
		err := suite.svc.UpdateLastSeen(tt.userID)
		assert.Equal(suite.T(), tt.wantErr, err)

		if tt.wantLS {
			user, _ := suite.svc.users.FindByID(tt.userID)
			assert.Equal(suite.T(), tt.wantLS, user.lastSeen.After(now))
		}
	}
}

func (suite *ServiceTestSuite) TestRegisterNewUser_AssignsUserAHashedPassword() {
	user, err := suite.svc.users.FindByID(suite.userID)

	assert.Nil(suite.T(), err)
	assert.True(suite.T(), checkPasswordHash(user.password, "password"))
}

func (suite *ServiceTestSuite) TestNewService() {
	users := NewUserRepository()
	posts := NewPostRepository()
	svc := NewService(users, posts)

	s := svc.(*service)

	assert.Equal(suite.T(), users, s.users)
	assert.Equal(suite.T(), posts, s.posts)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
