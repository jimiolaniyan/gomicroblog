package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestService_RegisterAccount(t *testing.T) {
	now := time.Now().UTC()
	accounts := NewAccountRepository()
	svc := NewService(accounts)

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
			assert.Equal(t, tt.wantCreatedAt, acc.CreatedAt.After(now))
			assert.True(t, hashMatchesPassword(acc.Credentials.Password, "password"))
		}
	}
}
