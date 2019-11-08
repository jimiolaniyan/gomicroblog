package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	u := &Account{Credentials: Credentials{Username: "user", Email: "e@m.co"}}
	longName := "long_name_that_exceeds_24_characters_should_not_be_allowed"

	tests := []struct {
		username, email string
		wantErr         error
		wantAcc         *Account
	}{
		{wantErr: ErrInvalidUsername},
		{username: longName, wantErr: ErrInvalidUsername},
		{username: "user name with space", wantErr: ErrInvalidUsername},
		{username: "user_name_with_@", wantErr: ErrInvalidUsername},
		{username: "username", wantErr: ErrInvalidEmail},
		{username: "username", email: "email", wantErr: ErrInvalidEmail},
		{username: "username", email: "email@sdf", wantErr: ErrInvalidEmail},
		{username: "user", email: "e@m.co", wantAcc: u},
	}

	for _, tt := range tests {
		user, err := NewAccount(tt.username, tt.email)
		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantAcc, user)
	}
}
