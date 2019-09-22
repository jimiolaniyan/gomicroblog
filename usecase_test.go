package gomicroblog

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignupReturnsErrorForEmptyUsername(t *testing.T) {
	user, err := Register("", "", "")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestSignupReturnsErrorForPassWordBelowEightCharacters(t *testing.T) {
	user, err := Register("username", "passwor", "")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestSignupReturnsErrorForEmptyPassword(t *testing.T) {
	user, err := Register("username", "", "")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestSignupReturnsErrorForEmptyEmail(t *testing.T) {
	user, err := Register("username", "password", "")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestSignupReturnsErrorForInvalidEmail(t *testing.T) {
	user, err := Register("username", "password", "email")

	assert.Error(t, err)
	assert.Nil(t, user)
}
