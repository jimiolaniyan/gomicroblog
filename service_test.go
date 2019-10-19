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
		{description: "ExistingUsername", req: &registerUserRequest{"username", "password1", "b@c"}, wantErr: ErrExistingUsername},
		{description: "ExistingEmail", req: &registerUserRequest{"username2", "password1", "a@b"}, wantErr: ErrExistingEmail},
		{description: "InvalidPassword", req: &registerUserRequest{"username2", "passwod", "b@c"}, wantErr: ErrInvalidPassword},
		{description: "ValidCredentials", req: &registerUserRequest{"username2", "password", "b@c.com"}, wantValidID: true, wantCreatedAt: true, wantLastSeen: true, wantErr: nil},
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
		{username: ts.req.Username, wantErr: nil, wantPostsLen: 1},
	}

	for _, tt := range tests {
		posts, err := ts.svc.GetUserPosts(tt.username)
		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantPostsLen, len(posts))
	}
}

func (ts *ServiceTestSuite) TestService_GetProfile() {
	av := avatar(ts.req.Email)
	u := ts.req.Username

	tests := []struct {
		username           string
		wantErr            error
		wantUN, wantAvatar string
		wantID             ID
		wantJ, wantLS      bool
	}{
		{username: "", wantErr: ErrInvalidUsername, wantUN: ""},
		{username: "void", wantErr: ErrNotFound, wantUN: ""},
		{username: u, wantErr: nil, wantUN: u, wantAvatar: av, wantJ: true, wantLS: true, wantID: ts.user.ID},
	}

	for _, tt := range tests {
		p, err := ts.svc.GetProfile(tt.username)

		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantUN, p.Username)
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
	// get a temporary user so we don't update the user in the suite
	tempUser := *ts.user
	tempUser.ID = nextID()
	origUN := "newUsername"
	tempUser.username = origUN
	bio := tempUser.bio
	err := ts.svc.users.Store(&tempUser)
	assert.Nil(ts.T(), err)

	b := "My new bio"
	longBio := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut " +
		"labore et dolore magna aliqua. Ut enim ad minim h"

	emptyUsernameReq := editProfileRequest{Username: new(string)}

	u := "U"

	tests := []struct {
		id      ID
		req     editProfileRequest
		wantErr error
		wantBio string
		wantUN  string
	}{
		{wantErr: ErrInvalidID},
		{req: emptyUsernameReq, wantErr: ErrInvalidID},
		{id: nextID(), req: emptyUsernameReq, wantErr: ErrNotFound},
		{id: tempUser.ID, req: emptyUsernameReq, wantErr: ErrInvalidUsername},
		{id: tempUser.ID, req: editProfileRequest{Username: &ts.req.Username}, wantErr: ErrExistingUsername},
		{id: tempUser.ID, req: editProfileRequest{Username: &origUN}, wantErr: nil, wantUN: origUN},
		{id: tempUser.ID, req: editProfileRequest{Username: &u}, wantErr: nil, wantUN: u, wantBio: bio},
		{id: tempUser.ID, req: editProfileRequest{Bio: &longBio}, wantErr: ErrBioTooLong, wantBio: bio, wantUN: origUN},
		{id: tempUser.ID, req: editProfileRequest{Bio: new(string)}, wantErr: nil, wantBio: "", wantUN: origUN},
		{id: tempUser.ID, req: editProfileRequest{Bio: &b}, wantErr: nil, wantBio: b, wantUN: origUN},
		{id: tempUser.ID, req: editProfileRequest{Username: &u, Bio: &b}, wantErr: nil, wantBio: b, wantUN: u},
	}

	for _, tt := range tests {
		err := ts.svc.EditProfile(tt.id, tt.req)
		assert.Equal(ts.T(), tt.wantErr, err)

		if err == nil {
			assert.Equal(ts.T(), tt.wantUN, tempUser.username)
			assert.Equal(ts.T(), tt.wantBio, tempUser.bio)
		}

		// reset
		tempUser.username = origUN
		tempUser.bio = bio
	}

	_ = ts.svc.users.Delete(tempUser.ID)
}

func (ts *ServiceTestSuite) TestCreateRelationshipFor() {
	u1 := duplicateUser(ts.svc, *ts.user, "user1")
	u2 := duplicateUser(ts.svc, *ts.user, "user2")

	tests := []struct {
		id       ID
		username string
		wantErr  error
		wantF    bool
		wantLen  int
	}{
		{id: "invalid", wantErr: ErrInvalidID},
		{id: nextID(), wantErr: ErrInvalidUsername},
		{id: nextID(), username: "nonexistent", wantErr: ErrNotFound},
		{id: u1.ID, username: u2.username, wantErr: nil, wantF: true, wantLen: 1},
		{id: u1.ID, username: u2.username, wantErr: nil, wantF: true, wantLen: 1},
	}

	for _, tt := range tests {
		err := ts.svc.CreateRelationshipFor(tt.id, tt.username)

		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantF, u1.IsFollowing(u2))
		assert.Equal(ts.T(), tt.wantLen, len(u1.Friends))
		assert.Equal(ts.T(), tt.wantLen, len(u2.Followers))
	}

	// clean up
	_ = ts.svc.users.Delete(u1.ID)
	_ = ts.svc.users.Delete(u2.ID)
}

func (ts *ServiceTestSuite) TestRelationships() {
	u1 := duplicateUser(ts.svc, *ts.user, "u1")
	u2 := duplicateUser(ts.svc, *ts.user, "u2")
	u1.Follow(u2)

	tests := []struct {
		username           string
		wantErr            error
		wantFriendsCount   int
		wantFollowersCount int
	}{
		{wantErr: ErrInvalidUsername},
		{username: "nonexistent", wantErr: ErrNotFound},
		{username: ts.req.Username, wantErr: nil},
		{username: u1.username, wantErr: nil, wantFriendsCount: 1},
		{username: u2.username, wantErr: nil, wantFollowersCount: 1},
	}

	for _, tt := range tests {
		friends, err := ts.svc.GetUserFriends(tt.username)
		followers, err := ts.svc.GetUserFollowers(tt.username)

		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantFriendsCount, len(friends))
		assert.Equal(ts.T(), tt.wantFollowersCount, len(followers))
	}

	// clean up
	_ = ts.svc.users.Delete(u1.ID)
	_ = ts.svc.users.Delete(u2.ID)
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
