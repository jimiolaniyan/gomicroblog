package gomicroblog

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeRequest(t *testing.T) {
	request := `
		{
			"username":"jimi",
			"password":"password1",
			"email":"a@b"
		}
`
	r := httptest.NewRequest("POST", "/users/v1/new", strings.NewReader(request))
	var d registerUserRequest
	body, err := decodeRequest(d, r)

	assert.Nil(t, err)
	assert.Equal(t, "jimi", body.Username)
	assert.Equal(t, "password1", body.Password)
	assert.Equal(t, "a@b", body.Email)
}

func TestDecodeRequestReturnsErrorForInvalidRequest(t *testing.T) {
	request := `name`
	r := httptest.NewRequest("POST", "/users/v1/new", strings.NewReader(request))

	var d registerUserRequest
	_, err := decodeRequest(d, r)

	assert.NotNil(t, err)
}

func TestHandlerGeneratesRequestModel(t *testing.T) {

}

func TestHandlerInvokesServiceWithRequestModelAndResponseGateway(t *testing.T) {

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
