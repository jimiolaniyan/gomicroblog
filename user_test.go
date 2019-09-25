package gomicroblog

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUser_ReturnsErrorForEmptyUsername(t *testing.T) {
	user, err := NewUser("", "")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestHashPassword_ReturnsCorrectHash(t *testing.T) {
	p := "password"
	hash, err := hashPassword(p)

	assert.Nil(t, err)
	assert.True(t, checkPasswordHash(hash, p))
}

func TestNewUser_ReturnsErrorForEmptyEmail(t *testing.T) {
	user, err := NewUser("username", "")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestNewUser_ReturnsErrorForInvalidEmail(t *testing.T) {
	user, err := NewUser("username", "email")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestNewUser_ReturnsUserWithSpecifiedArguments(t *testing.T) {
	username := "user"
	email := "email@email.com"

	user, err := NewUser(username, email)

	assert.Nil(t, err)
	assert.Equal(t, username, user.username)
	assert.Equal(t, email, user.email)
}
