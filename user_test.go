package blog

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

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
	u1 := &User{Username: "rand1", Email: "rand1@r.co"}
	u2 := &User{Username: "rand2", Email: "rand2@r.co"}
	u3 := &User{Username: "rand3", Email: "rand3@r.co"}
	u4 := &User{Username: "rand4", Email: "rand4@r.co"}

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
