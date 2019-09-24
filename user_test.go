package gomicroblog

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUser_ReturnsErrorForEmptyUsername(t *testing.T) {
	user, err := NewUser("", "", "")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestNewUser_ReturnsErrorForPassWordBelowEightCharacters(t *testing.T) {
	user, err := NewUser("username", "passwor", "")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestNewUser_ReturnsErrorForEmptyPassword(t *testing.T) {
	user, err := NewUser("username", "", "")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestNewUser_ReturnsErrorForEmptyEmail(t *testing.T) {
	user, err := NewUser("username", "password", "")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestNewUser_ReturnsErrorForInvalidEmail(t *testing.T) {
	user, err := NewUser("username", "password", "email")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestNewUser_ReturnsUserWithSpecifiedArguments(t *testing.T) {
	username := "user"
	password := "password"
	email := "email@email.com"

	user, err := NewUser(username, password, email)

	assert.Nil(t, err)
	assert.Equal(t, username, user.username)
	assert.Equal(t, password, user.password)
	assert.Equal(t, email, user.email)
}
