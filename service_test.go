package gomicroblog

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ServiceTestSuite struct {
	suite.Suite
	svc service
	req registerUserRequest
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.svc = service{users: NewUserRepository()}
	suite.req = registerUserRequest{"username", "password", "a@b"}
}

func TestService_RegisterNewUser(t *testing.T) {
	svc := service{users: NewUserRepository()}
	req := registerUserRequest{"username", "password", "a@b"}

	tests := []struct {
		description string
		req         *registerUserRequest
		wantValidID bool
		wantErr     error
	}{
		{
			"ExistingUsername",
			&registerUserRequest{"username", "password1", "b@c"},
			false,
			ErrExistingUsername,
		},
		{
			"ExistingEmail",
			&registerUserRequest{"username2", "password1", "a@b"},
			false,
			ErrExistingEmail,
		},
		{
			"InvalidPassword",
			&registerUserRequest{"username2", "passwod", "b@c"},
			false,
			ErrInvalidPassword,
		},
		{
			"ValidCredentials",
			&registerUserRequest{"username2", "password", "b@c.com"},
			true,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.description), func(t *testing.T) {
			_, err := svc.RegisterNewUser(req)
			userID, err := svc.RegisterNewUser(*tt.req)

			assert.Equal(t, IsValidID(string(userID)), tt.wantValidID)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateUser(t *testing.T) {
	svc := service{NewUserRepository()}
	_, err := svc.RegisterNewUser(registerUserRequest{"user", "password", "a@b.com"})
	assert.Nil(t, err)

	tests := []struct {
		username, password string
		wantErr            error
		wantValidID        bool
	}{
		{"", "", ErrInvalidCredentials, false},
		{"user", "jaiu", ErrInvalidCredentials, false},
		{"nonexistent", "password", ErrInvalidCredentials, false},
		{"user", "incorrect", ErrInvalidCredentials, false},
		{"user", "password", nil, true},
	}

	for _, tt := range tests {
		req := validateUserRequest{tt.username, tt.password}

		userID, err := svc.ValidateUser(req)

		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantValidID, IsValidID(string(userID)))
	}

}

func (suite *ServiceTestSuite) TestRegisterNewUser_AssignsUserAHashedPassword() {
	userID, _ := suite.svc.RegisterNewUser(suite.req)

	user, err := suite.svc.users.FindByID(userID)

	assert.Nil(suite.T(), err)
	assert.True(suite.T(), checkPasswordHash(user.password, "password"))
}

func (suite *ServiceTestSuite) TestNewService() {
	users := NewUserRepository()
	svc := NewService(users)
	s := svc.(*service)

	assert.Equal(suite.T(), users, s.users)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
