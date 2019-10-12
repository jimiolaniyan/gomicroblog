package gomicroblog

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
	tests := []struct {
		username, email string
		wantErr         error
		wantUser        *user
	}{
		{"", "", ErrInvalidUsername, nil},
		{"username", "", ErrInvalidEmail, nil},
		{"username", "email", ErrInvalidEmail, nil},
		{"user", "email@email.com", nil, &user{username: "user", email: "email@email.com"}},
	}

	for _, tt := range tests {
		user, err := NewUser(tt.username, tt.email)
		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantUser, user)
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
