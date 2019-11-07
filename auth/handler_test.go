package auth

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeRequest(t *testing.T) {
	registerReq := `{"username": "u", "email": "a@b.com", "password": "password1"}`
	body := ioutil.NopCloser(strings.NewReader(registerReq))
	r := registerAccountRequest{"u", "a@b.com", "password1"}

	req, err := decodeRegisterAccountRequest(body)

	assert.NoError(t, err)
	assert.Equal(t, r, req)
}

var errNil = errors.New("")

func TestRegisterNewUserHandler(t *testing.T) {
	invalidUsernameReq := `{"username": "", "password": "pass"}`
	invalidPassReq := `{"username": "u", "email": "a@b.com", "password": "pass"}`
	invalidEmailReq := `{"username": "u", "email": "a@bcom", "password": "password"}`
	registerReq := `{"username":"u", "email":"a@b.com", "password":"password1"}`
	existingUserReq := `{"username": "u", "email": "a@b.co", "password": "password"}`
	existingEmailReq := `{"username": "u2", "email": "a@b.com", "password": "password"}`

	accounts := NewAccountRepository()
	b := &eventsSpy{}
	svc := NewService(accounts, b)

	tests := []struct {
		req          string
		wantCode     int
		wantID       bool
		wantErr      error
		wantLocation string
	}{
		{req: `invalid request`, wantCode: http.StatusBadRequest, wantErr: errNil},
		{req: invalidUsernameReq, wantCode: http.StatusUnprocessableEntity, wantErr: ErrInvalidUsername},
		{req: invalidPassReq, wantCode: http.StatusUnprocessableEntity, wantErr: ErrInvalidPassword},
		{req: invalidEmailReq, wantCode: http.StatusUnprocessableEntity, wantErr: ErrInvalidEmail},
		{req: registerReq, wantCode: http.StatusCreated, wantID: true, wantErr: errNil, wantLocation: "/auth/v1/accounts"},
		{req: existingUserReq, wantCode: http.StatusConflict, wantErr: ErrExistingUsername},
		{req: existingEmailReq, wantCode: http.StatusConflict, wantErr: ErrExistingEmail},
	}

	for _, tt := range tests {
		r, _ := http.NewRequest(http.MethodPost, "/auth/v1/accounts", strings.NewReader(tt.req))

		w := httptest.NewRecorder()
		handler := http.NewServeMux()
		handler.Handle("/auth/v1/accounts", RegisterAccountHandler(svc))
		handler.ServeHTTP(w, r)

		var res struct {
			ID  ID     `json:"id,omitempty"`
			Err string `json:"error,omitempty"`
		}

		_ = json.NewDecoder(w.Body).Decode(&res)
		assert.Equal(t, tt.wantCode, w.Code)
		assert.Equal(t, tt.wantErr.Error(), res.Err)
		assert.Equal(t, isValidID(string(res.ID)), tt.wantID)
		assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
		assert.True(t, strings.HasPrefix(w.Header().Get("Location"), tt.wantLocation))
	}
}
