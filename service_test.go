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
		req          *registerUserRequest
		wantIDMinLen int
		wantErr      error
	}{
		{
			&registerUserRequest{
				"username",
				"password1",
				"b@c",
			},
			-1,
			ErrExistingUsername,
		},
		{
			&registerUserRequest{
				"username2",
				"password1",
				"a@b",
			},
			-1,
			ErrExistingEmail,
		},
		{
			&registerUserRequest{
				"username2",
				"passwod",
				"b@c",
			},
			-1,
			ErrInvalidPassword,
		},
		{
			&registerUserRequest{
				"username2",
				"password",
				"b@c.com",
			},
			5,
			nil,
		},
	}

	for kk, tt := range tests {
		t.Run(fmt.Sprintf("%d", kk), func(t *testing.T) {
			_, err := svc.RegisterNewUser(req)
			userID, err := svc.RegisterNewUser(*tt.req)

			assert.Greater(t, len(string(userID)), tt.wantIDMinLen)
			assert.Equal(t, tt.wantErr, err)
		})
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
