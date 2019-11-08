package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/assert"
)

type ServiceTestSuite struct {
	suite.Suite
	svc      Service
	accounts Repository
	username string
	id       ID
}

func TestService_RegisterAccount(t *testing.T) {
	now := time.Now().UTC()
	accounts := NewAccountRepository()
	spy := &eventsSpy{}
	svc := NewService(accounts, spy)

	tests := []struct {
		req                            registerAccountRequest
		wantErr                        error
		wantID, wantCreatedAt, wantAcc bool
	}{
		{req: registerAccountRequest{"u", "b@c.com", "invalid"}, wantErr: ErrInvalidPassword},
		{req: registerAccountRequest{"u", "b@c.com", "password"}, wantID: true, wantCreatedAt: true, wantAcc: true},
		{req: registerAccountRequest{"u", "b@c.com", "password"}, wantErr: ErrExistingUsername},
		{req: registerAccountRequest{"u2", "b@c.com", "password"}, wantErr: ErrExistingEmail},
	}

	for _, tt := range tests {
		id, err := svc.RegisterAccount(tt.req)

		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantID, isValidID(string(id)))

		acc, err := accounts.FindByID(id)
		if tt.wantAcc {
			assert.NoError(t, err)
			assert.Equal(t, spy.id, string(acc.ID))
			assert.Equal(t, spy.username, acc.Credentials.Username)
			assert.Equal(t, spy.email, acc.Credentials.Email)
			assert.Equal(t, tt.wantCreatedAt, acc.CreatedAt.After(now))
			assert.True(t, hashMatchesPassword(acc.Credentials.Password, "password"))
		}
	}
}

type eventsSpy struct {
	id, username, email string
}

func TestService_ValidateUser(t *testing.T) {
	accounts := NewAccountRepository()
	svc := NewService(accounts, &eventsSpy{})
	_, _ = svc.RegisterAccount(registerAccountRequest{"test", "test@test.com", "password"})

	tests := []struct {
		username, password string
		wantErr            error
		wantValidID        bool
	}{
		{wantErr: ErrInvalidCredentials},
		{username: "user", password: "pass", wantErr: ErrInvalidCredentials},
		{username: "nonexistent", password: "password", wantErr: ErrInvalidCredentials},
		{username: "test", password: "incorrect", wantErr: ErrInvalidCredentials},
		{username: "test", password: "password", wantValidID: true},
	}

	for _, tt := range tests {
		req := validateCredentialsRequest{tt.username, tt.password}

		userID, err := svc.ValidateCredentials(req)

		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantValidID, isValidID(string(userID)))
	}
}

func (a *eventsSpy) AccountCreated(id string, username string, email string) {
	a.id = id
	a.username = username
	a.email = email
}
