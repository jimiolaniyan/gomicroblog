package gomicroblog

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
		{"", "", ErrEmptyUserName, nil},
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
