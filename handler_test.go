package gomicroblog

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type HandlerTestSuite struct {
	suite.Suite
	registerReq, loginReq string
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.registerReq = `{ "username": "jimi", "password": "password1", "email": "test@tester.test" }`
	suite.loginReq = `{"username": "jimi", "password": "password1"}`
}

func (suite *HandlerTestSuite) TestDecodeRequest() {
	tests := []struct {
		r       string
		decoder func(closer io.ReadCloser) (interface{}, error)
		wantErr error
		wantReq interface{}
	}{
		{suite.registerReq, decodeRegisterUserRequest, nil, registerUserRequest{"jimi", "password1", "test@tester.test"}},
		{suite.loginReq, decodeValidateUserRequest, nil, validateUserRequest{"jimi", "password1"}},
	}

	for _, tt := range tests {
		body := ioutil.NopCloser(strings.NewReader(tt.r))
		req, err := tt.decoder(body)
		assert.Equal(suite.T(), tt.wantErr, err)
		assert.Equal(suite.T(), tt.wantReq, req)
	}
}

func (suite *HandlerTestSuite) TestHandlerInvokesServiceWithRequest() {
	r, err := http.NewRequest(http.MethodPost, "/v1/users/new", strings.NewReader(suite.registerReq))
	assert.Nil(suite.T(), err)

	svc := &ServiceSpy{}

	w := httptest.NewRecorder()
	handler := http.NewServeMux()
	handler.Handle("/v1/users/new", RegisterUserHandler(svc))
	handler.ServeHTTP(w, r)

	assert.True(suite.T(), svc.registerNewUserWasCalled)
	assert.Equal(suite.T(), "jimi", svc.request.Username)
	assert.Equal(suite.T(), "password1", svc.request.Password)
	assert.Equal(suite.T(), "test@tester.test", svc.request.Email)
}

func TestHandlerResponses(t *testing.T) {
	svc := &service{users: NewUserRepository()}
	url := "/v1/users/new"
	registerHandler := RegisterUserHandler(svc)
	registerReq := `
		{
			"username":"jimi",
			"password":"password1",
			"email":"test@tester.test"
		}
`
	tests := []struct {
		method, req  string
		wantCode     int
		wantValidID  bool
		wantErr      error
		wantLocation string
		testExisting bool
	}{
		{
			http.MethodPost,
			registerReq,
			http.StatusCreated,
			true,
			errors.New(""),
			"/v1/users",
			false,
		},
		{
			http.MethodPost,
			`invalid request`,
			http.StatusBadRequest,
			false,
			errors.New(""),
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "", "password": "pass"}`,
			http.StatusUnprocessableEntity,
			false,
			ErrInvalidUsername,
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "username", "password": "pass", "email": "a@b.com"}`,
			http.StatusUnprocessableEntity,
			false,
			ErrInvalidPassword,
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "username", "password": "password", "email": "ab.com"}`,
			http.StatusUnprocessableEntity,
			false,
			ErrInvalidEmail,
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "jimi", "password": "password", "email": "a@b.com"}`,
			http.StatusConflict,
			false,
			ErrExistingUsername,
			"",
			true,
		},
		{
			http.MethodPost,
			`{"username": "username", "password": "password", "email": "test@tester.test"}`,
			http.StatusConflict,
			false,
			ErrExistingEmail,
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.wantCode), func(t *testing.T) {
			if tt.testExisting {
				ur := registerUserRequest{}
				req := registerReq
				_ = json.NewDecoder(strings.NewReader(req)).Decode(&ur)
				user, _ := NewUser(ur.Username, ur.Email)
				_ = svc.users.Store(user)
			}

			r, err := http.NewRequest(tt.method, url, strings.NewReader(tt.req))
			assert.Nil(t, err)

			w := httptest.NewRecorder()
			handler := http.NewServeMux()
			handler.Handle(url, registerHandler)
			handler.ServeHTTP(w, r)

			var res struct {
				ID  ID     `json:"id,omitempty"`
				Err string `json:"error,omitempty"`
			}

			_ = json.NewDecoder(w.Body).Decode(&res)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantErr.Error(), res.Err)
			assert.Equal(t, IsValidID(string(res.ID)), tt.wantValidID)
			assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
			assert.True(t, strings.HasPrefix(w.Header().Get("Location"), tt.wantLocation))
		})
	}
}

func TestLoginHandler(t *testing.T) {
	svc := NewService(NewUserRepository())
	_, _ = svc.RegisterNewUser(registerUserRequest{"user", "password", "a@b.com"})

	tests := []struct {
		description, req string
		wantCode         int
	}{
		{"BadRequest", `invalid request`, http.StatusBadRequest},
		{"NonExistentUser", `{"username": "nonexistent", "password": "password"}`, http.StatusUnauthorized},
		{"ExistingUserWithInvalidPassword", `{"username": "user", "password": "anInvalid"}`, http.StatusUnauthorized},
		{"ExistingUserWithValidPassword", `{"username": "user", "password": "password"}`, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(tt.req))
			assert.Nil(t, err)

			w := httptest.NewRecorder()
			mux := http.NewServeMux()
			mux.Handle("/v1/auth/login", LoginHandler(svc))
			mux.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}

}

type ServiceSpy struct {
	registerNewUserWasCalled bool
	request                  registerUserRequest
}

func (s *ServiceSpy) ValidateUser(req validateUserRequest) (ID, error) {
	return "", nil
}

func (s *ServiceSpy) RegisterNewUser(req registerUserRequest) (ID, error) {
	s.registerNewUserWasCalled = true
	s.request = req
	return "", nil
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
