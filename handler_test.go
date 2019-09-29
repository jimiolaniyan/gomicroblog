package gomicroblog

import (
	"encoding/json"
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
	r := httptest.NewRequest("POST", "/users/v1/new", strings.NewReader(suite.registerReq))

	body, err := decodeRegisterUserRequest(r)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "jimi", body.Username)
	assert.Equal(suite.T(), "password1", body.Password)
	assert.Equal(suite.T(), "test@tester.test", body.Email)
}

func (suite *HandlerTestSuite) TestDecodeRequestReturnsErrorForInvalidRequest() {
	request := `name`
	r := httptest.NewRequest("POST", "/users/v1/new", strings.NewReader(request))

	_, err := decodeRegisterUserRequest(r)

	assert.NotNil(suite.T(), err)
}

func (suite *HandlerTestSuite) TestHandlerInvokesServiceWithRequest() {
	r, err := http.NewRequest("POST", "/users/v1/new", strings.NewReader(suite.registerReq))
	assert.Nil(suite.T(), err)

	svc := &ServiceSpy{}

	w := httptest.NewRecorder()
	handler := http.NewServeMux()
	handler.Handle("/users/v1/new", RegisterUserHandler(svc))
	handler.ServeHTTP(w, r)

	assert.True(suite.T(), svc.registerNewUserWasCalled)
	assert.Equal(suite.T(), "jimi", svc.request.Username)
	assert.Equal(suite.T(), "password1", svc.request.Password)
	assert.Equal(suite.T(), "test@tester.test", svc.request.Email)
}

func (suite *HandlerTestSuite) TestHandlerReturnsEncodedResponse() {
	r, err := http.NewRequest("POST", "/users/v1/new", strings.NewReader(suite.registerReq))
	assert.Nil(suite.T(), err)

	svc := &service{users: NewUserRepository()}
	w := httptest.NewRecorder()
	handler := http.NewServeMux()
	handler.Handle("/users/v1/new", RegisterUserHandler(svc))
	handler.ServeHTTP(w, r)

	var res struct {
		ID ID `json:"id"`
	}

	json.NewDecoder(w.Body).Decode(&res)
	fmt.Println(res.ID)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Greater(suite.T(), len(res.ID), 3)
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
