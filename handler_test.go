package gomicroblog

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type HandlerTestSuite struct {
	suite.Suite
	registerReq string
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.registerReq = `
		{
			"username":"jimi",
			"password":"password1",
			"email":"test@tester.test"
		}
`
}

func (suite *HandlerTestSuite) TestDecodeRequest() {
	r := httptest.NewRequest(http.MethodPost, "/v1/users/new", strings.NewReader(suite.registerReq))

	body, err := decodeRegisterUserRequest(r)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "jimi", body.Username)
	assert.Equal(suite.T(), "password1", body.Password)
	assert.Equal(suite.T(), "test@tester.test", body.Email)
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
		method, req          string
		wantCode, wantMinLen int
		wantErr              error
		wantLocation         string
		testExisting         bool
	}{
		{
			http.MethodPost,
			registerReq,
			http.StatusCreated,
			3,
			errors.New(""),
			"/v1/users",
			false,
		},
		{
			http.MethodPost,
			`invalid request`,
			http.StatusBadRequest,
			-1,
			errors.New(""),
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "", "password": "pass"}`,
			http.StatusUnprocessableEntity,
			-1,
			ErrEmptyUserName,
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "username", "password": "pass", "email": "a@b.com"}`,
			http.StatusUnprocessableEntity,
			-1,
			ErrInvalidPassword,
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "username", "password": "password", "email": "ab.com"}`,
			http.StatusUnprocessableEntity,
			-1,
			ErrInvalidEmail,
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "jimi", "password": "password", "email": "a@b.com"}`,
			http.StatusConflict,
			-1,
			ErrExistingUsername,
			"",
			true,
		},
		{
			http.MethodPost,
			`{"username": "username", "password": "password", "email": "test@tester.test"}`,
			http.StatusConflict,
			-1,
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

			json.NewDecoder(w.Body).Decode(&res)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantErr.Error(), res.Err)
			assert.Greater(t, len(res.ID), tt.wantMinLen)
			assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
			assert.True(t, strings.HasPrefix(w.Header().Get("Location"), tt.wantLocation))
		})
	}
}

type ServiceSpy struct {
	registerNewUserWasCalled bool
	request                  registerUserRequest
}

func (s *ServiceSpy) RegisterNewUser(req registerUserRequest) (ID, error) {
	s.registerNewUserWasCalled = true
	s.request = req
	return "", nil
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
