package gomicroblog

import (
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

func (suite *HandlerTestSuite) TestHandlerInvokesServiceWithRequestAndResponder() {

	r, err := http.NewRequest("POST", "/users/v1/new", strings.NewReader(suite.registerReq))
	assert.Nil(suite.T(), err)

	svc := &ServiceSpy{}
	responderSpy := &ResponderSpy{}

	w := httptest.NewRecorder()
	handler := http.NewServeMux()
	handler.Handle("/users/v1/new", RegisterUserHandler(svc, responderSpy))
	handler.ServeHTTP(w, r)

	assert.True(suite.T(), svc.registerNewUserWasCalled)
	assert.Equal(suite.T(), "jimi", svc.request.Username)
	assert.Equal(suite.T(), "password1", svc.request.Password)
	assert.Equal(suite.T(), "test@tester.test", svc.request.Email)
	assert.Equal(suite.T(), responderSpy, svc.responder)
}

//func TestServiceCallsFormatterWithResponseModel(t *testing.T) {
//
//}
//
//func TestFormatterGeneratesFormattedModel(t *testing.T) {
//
//}

func TestHandlerSendsFormattedModelToEncoder(t *testing.T) {

}

type ServiceSpy struct {
	registerNewUserWasCalled bool
	request                  registerUserRequest
	responder                Responder
}

func (s *ServiceSpy) RegisterNewUser(req registerUserRequest, res Responder) {
	s.registerNewUserWasCalled = true
	s.request = req
	s.responder = res
	return
}

type ResponderSpy struct {
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
