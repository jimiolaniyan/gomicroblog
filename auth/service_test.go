package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestService_RegisterAccount(t *testing.T) {
	now := time.Now().UTC()
	accounts := NewAccountRepository()
	spy := &accountEventsSpy{}
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

type accountEventsSpy struct {
	id, username, email string
}

func (a *accountEventsSpy) AccountCreated(accID string, username string, email string) {
	a.id = accID
	a.username = username
	a.email = email
}
