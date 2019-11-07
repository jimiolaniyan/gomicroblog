package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

func TestDecodeLoginRequest(t *testing.T) {
	registerReq := `{"username": "u", "password": "password"}`
	body := ioutil.NopCloser(strings.NewReader(registerReq))
	r := validateCredentialsRequest{"u", "password"}

	d, err := decodeLoginRequest(body)

	assert.NoError(t, err)
	assert.Equal(t, r, d)
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

func TestLoginHandler(t *testing.T) {
	svc := NewService(NewAccountRepository(), &eventsSpy{})
	id, _ := svc.RegisterAccount(registerAccountRequest{"test", "t@t.com", "password"})
	validClaims := fmt.Sprintf("{\"iss\":\"auth\",\"sub\":\"%s\"}", id)
	tests := []struct {
		req                    string
		wantCode, wantTokenLen int
		wantClaims             string
	}{
		{req: `invalid request`, wantCode: http.StatusBadRequest, wantTokenLen: 1},
		{req: `{"username": "nonexistent", "password": "password"}`, wantCode: http.StatusUnauthorized, wantTokenLen: 1},
		{req: `{"username": "test", "password": "anInvalid"}`, wantCode: http.StatusUnauthorized, wantTokenLen: 1},
		{req: `{"username": "test", "password": "password"}`, wantCode: http.StatusOK, wantTokenLen: 3, wantClaims: validClaims},
	}

	for _, tt := range tests {
		req, err := http.NewRequest(http.MethodPost, "/v1/sessions", strings.NewReader(tt.req))
		assert.Nil(t, err)

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.Handle("/v1/sessions", LoginHandler(svc))
		mux.ServeHTTP(w, req)

		var res struct {
			Token string `json:"token,omitempty"`
			Error string `json:"error,omitempty"`
		}

		_ = json.NewDecoder(w.Body).Decode(&res)
		assert.Equal(t, tt.wantCode, w.Code)
		parts := strings.Split(res.Token, ".")
		assert.Equal(t, len(parts), tt.wantTokenLen)

		if len(parts) > 2 {
			claim, err := base64.RawStdEncoding.DecodeString(parts[1])
			assert.Nil(t, err)
			assert.Equal(t, tt.wantClaims, string(claim))
		}
	}
}
