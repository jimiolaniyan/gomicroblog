package blog

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

func TestHashPassword_ReturnsCorrectHash(t *testing.T) {
	p := "password"
	hash, err := hashPassword(p)

	assert.Nil(t, err)
	assert.True(t, checkPasswordHash(hash, p))
}

func TestNewUser(t *testing.T) {
	u := &User{Username: "user", Email: "e@m.co"}

	tests := []struct {
		username, email string
		wantErr         error
		wantUser        *User
	}{
		{wantErr: ErrInvalidUsername},
		{username: "username", wantErr: ErrInvalidEmail},
		{username: "username", email: "email", wantErr: ErrInvalidEmail},
		{username: "user", email: "e@m.co", wantUser: u},
	}

	for _, tt := range tests {
		user, err := NewUser(tt.username, tt.email)
		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantUser, user)

		if user != nil {
			assert.Equal(t, tt.wantUser.Friends, user.Friends)
			assert.Equal(t, tt.wantUser.Followers, user.Followers)
		}
	}
}

func TestValidID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{id: "", want: false},
		{id: "egegeg", want: false},
		{id: "egege84g9f8dw9f929d9", want: false},
		{id: xid.New().String(), want: true},
	}

	for _, tt := range tests {
		v := IsValidID(tt.id)
		if tt.want != v {
			t.Errorf("Test failed; want %t, got %t", tt.want, v)
		}
	}
}

func TestUser_FollowAndUnFollow(t *testing.T) {
	u1, _ := NewUser("rand1", "rand1@r.co")
	u2, _ := NewUser("rand2", "rand2@r.co")
	u3, _ := NewUser("rand3", "rand3@r.co")
	u4, _ := NewUser("rand4", "rand4@r.co")

	u1.ID = nextID()
	u2.ID = nextID()
	u3.ID = nextID()
	u4.ID = nextID()

	u1.Follow(u2)
	u1.Follow(u3)
	u1.Follow(u4)
	u2.Follow(u1)
	u2.Follow(u3)
	u4.Follow(u1)
	u4.Follow(u2)

	assert.Equal(t, 3, len(u1.Friends))
	assert.Equal(t, 2, len(u1.Followers))
	assert.Equal(t, 2, len(u2.Friends))
	assert.Equal(t, 2, len(u2.Followers))
	assert.Equal(t, 2, len(u4.Friends))

	u1.Unfollow(u2)
	u2.Unfollow(u3)
	u1.Unfollow(u4)
	u4.Unfollow(u1)
	u4.Unfollow(u2)

	assert.Equal(t, 1, len(u1.Friends))
	assert.Equal(t, 1, len(u1.Followers))
	assert.Equal(t, 1, len(u2.Friends))
	assert.Equal(t, 0, len(u2.Followers))
	assert.Equal(t, 0, len(u4.Friends))
	assert.Equal(t, 0, len(u4.Followers))
}
