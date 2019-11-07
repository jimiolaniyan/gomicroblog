package blog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	svc             service
	username, email string
	userID          ID
	user            *User
}

func (ts *ServiceTestSuite) TearDownTest() {
	ts.svc.posts = NewPostRepository()
}

func (ts *ServiceTestSuite) SetupSuite() {
	ts.svc = service{users: NewUserRepository(), posts: NewPostRepository()}
	ts.userID = nextID()
	ts.username = "username"
	ts.email = "a@b.con"
	ts.svc.CreateProfile(string(ts.userID), ts.username, ts.email)

	u, _ := ts.svc.users.FindByID(ts.userID)
	ts.user = u
}

func (ts *ServiceTestSuite) TestService_CreateProfile() {
	id := nextID()
	now := time.Now().UTC()
	tests := []struct {
		username, email string
		wantCreatedAt   bool
	}{
		{username: ts.username},
		{username: "new", email: ts.email},
		{username: "new", email: "n@e.co", wantCreatedAt: true},
	}

	for _, tt := range tests {
		ts.svc.CreateProfile(string(id), tt.username, tt.email)

		user, _ := ts.svc.users.FindByID(id)

		if user != nil {
			assert.Equal(ts.T(), tt.wantCreatedAt, user.CreatedAt.After(now))
		}
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
			assert.Equal(ts.T(), tt.body, post.Body)
			assert.Equal(ts.T(), tt.userID, post.Author.UserID)
			assert.Equal(ts.T(), tt.wantTS, post.Timestamp.After(now))
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
		{username: ts.username, wantErr: nil, wantPostsLen: 1},
	}

	for _, tt := range tests {
		posts, err := ts.svc.GetUserPosts(tt.username)
		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantPostsLen, len(posts))
	}
}

func (ts *ServiceTestSuite) TestService_GetProfile() {
	av := avatar(ts.email)
	u := ts.username

	tests := []struct {
		username           string
		wantErr            error
		wantUN, wantAvatar string
		wantID             ID
	}{
		{username: "", wantErr: ErrInvalidUsername, wantUN: ""},
		{username: "void", wantErr: ErrNotFound, wantUN: ""},
		{username: u, wantErr: nil, wantUN: u, wantAvatar: av, wantID: ts.user.ID},
	}

	for _, tt := range tests {
		p, err := ts.svc.GetProfile(tt.username)

		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantUN, p.Username)
		assert.Equal(ts.T(), tt.wantAvatar, p.Avatar)
		assert.Equal(ts.T(), tt.wantID, p.ID)

		if tt.wantErr == nil {
			assert.Equal(ts.T(), ts.user.CreatedAt, p.Joined)
			assert.Equal(ts.T(), ts.user.LastSeen, p.LastSeen)
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
			assert.Equal(ts.T(), tt.wantLS, ts.user.LastSeen.After(now))
		}
	}
}

func (ts *ServiceTestSuite) TestEditProfile() {
	// get a temporary user so we don't update the user in the suite
	tempUser := *ts.user
	tempUser.ID = nextID()
	origUN := "newUsername"
	tempUser.Username = origUN
	bio := tempUser.Bio
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
		{id: tempUser.ID, req: editProfileRequest{Username: &ts.username}, wantErr: ErrExistingUsername},
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
			assert.Equal(ts.T(), tt.wantUN, tempUser.Username)
			assert.Equal(ts.T(), tt.wantBio, tempUser.Bio)
		}

		// reset
		tempUser.Username = origUN
		tempUser.Bio = bio
	}

	_ = ts.svc.users.Delete(tempUser.ID)
}

func (ts *ServiceTestSuite) TestCreateRelationshipFor() {
	u1 := DuplicateUser(ts.svc.users, *ts.user, "user1")
	u2 := DuplicateUser(ts.svc.users, *ts.user, "user2")

	tests := []struct {
		id         ID
		username   string
		wantErr    error
		wantFollow bool
		wantLen    int
	}{
		{id: "invalid", wantErr: ErrInvalidID},
		{id: nextID(), wantErr: ErrInvalidUsername},
		{id: nextID(), username: "nonexistent", wantErr: ErrNotFound},
		{id: u1.ID, username: u1.Username, wantErr: ErrCantFollowSelf},
		{id: u1.ID, username: u2.Username, wantErr: nil, wantFollow: true, wantLen: 1},
		{id: u1.ID, username: u2.Username, wantErr: ErrAlreadyFollowing, wantFollow: true, wantLen: 1},
	}

	for _, tt := range tests {
		err := ts.svc.CreateRelationshipFor(tt.id, tt.username)

		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantFollow, u1.IsFollowing(u2))
		assert.Equal(ts.T(), tt.wantLen, len(u1.Friends))
		assert.Equal(ts.T(), tt.wantLen, len(u2.Followers))
	}

	// clean up
	_ = ts.svc.users.Delete(u1.ID)
	_ = ts.svc.users.Delete(u2.ID)
}

func (ts *ServiceTestSuite) TestRemoveRelationshipFor() {
	u2 := DuplicateUser(ts.svc.users, *ts.user, "abc")
	u1 := DuplicateUser(ts.svc.users, *ts.user, "xyz")
	u1.Follow(u2)

	tests := []struct {
		id         ID
		wantErr    error
		username   string
		wantFollow bool
		wantLen    int
	}{
		{id: "invalid", wantErr: ErrInvalidID},
		{id: nextID(), wantErr: ErrInvalidUsername},
		{id: nextID(), username: "nonexistent", wantErr: ErrNotFound},
		{id: u1.ID, username: u1.Username, wantErr: ErrCantUnFollowSelf},
		{id: u1.ID, username: u2.Username, wantErr: nil},
		{id: u1.ID, username: u2.Username, wantErr: ErrNotFollowing},
	}

	for _, tt := range tests {
		err := ts.svc.RemoveRelationshipFor(tt.id, tt.username)

		assert.Equal(ts.T(), tt.wantErr, err)

		if err == nil {
			assert.Equal(ts.T(), tt.wantFollow, u1.IsFollowing(u2))
			assert.Equal(ts.T(), tt.wantLen, len(u1.Friends))
			assert.Equal(ts.T(), tt.wantLen, len(u2.Followers))
		}
	}

	// clean up
	_ = ts.svc.users.Delete(u1.ID)
	_ = ts.svc.users.Delete(u2.ID)
}

func (ts *ServiceTestSuite) TestRelationships() {
	u1 := DuplicateUser(ts.svc.users, *ts.user, "u1")
	u2 := DuplicateUser(ts.svc.users, *ts.user, "u2")
	u1.Follow(u2)

	tests := []struct {
		username           string
		wantErr            error
		wantFriendsCount   int
		wantFollowersCount int
	}{
		{wantErr: ErrInvalidUsername},
		{username: "nonexistent", wantErr: ErrNotFound},
		{username: ts.username, wantErr: nil},
		{username: u1.Username, wantErr: nil, wantFriendsCount: 1},
		{username: u2.Username, wantErr: nil, wantFollowersCount: 1},
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

func (ts *ServiceTestSuite) TestGetTimeline() {
	u1 := DuplicateUser(ts.svc.users, *ts.user, "a1")
	u2 := DuplicateUser(ts.svc.users, *ts.user, "a2")
	u3 := DuplicateUser(ts.svc.users, *ts.user, "a3")
	u4 := DuplicateUser(ts.svc.users, *ts.user, "a4")
	u5 := DuplicateUser(ts.svc.users, *ts.user, "a5")

	u1.Follow(u2)
	u1.Follow(u3)
	u1.Follow(u4)
	u5.Follow(u4)

	_, _ = ts.svc.CreatePost(u2.ID, "p2")
	_, _ = ts.svc.CreatePost(u1.ID, "p3")
	_, _ = ts.svc.CreatePost(u2.ID, "p4")
	_, _ = ts.svc.CreatePost(u4.ID, "p5")

	tests := []struct {
		id          ID
		wantErr     error
		wantPostLen int
	}{
		{id: "invalid", wantErr: ErrInvalidID},
		{id: nextID(), wantErr: ErrNotFound},
		{id: u3.ID},
		{id: u2.ID, wantPostLen: 2},
		{id: u4.ID, wantPostLen: 1},
		{id: u5.ID, wantPostLen: 1},
		{id: u1.ID, wantPostLen: 4},
	}

	for _, tt := range tests {
		tl, err := ts.svc.GetTimeline(tt.id)

		assert.Equal(ts.T(), tt.wantErr, err)
		assert.Equal(ts.T(), tt.wantPostLen, len(tl))
	}

	// clean up
	_ = ts.svc.users.Delete(u1.ID)
	_ = ts.svc.users.Delete(u2.ID)
	_ = ts.svc.users.Delete(u3.ID)
	_ = ts.svc.users.Delete(u4.ID)
	_ = ts.svc.users.Delete(u5.ID)
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
