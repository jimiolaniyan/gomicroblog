package gomicroblog

import (
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

func (suite *ServiceTestSuite) TestService_RegisterNewUser() {
	tests := []struct {
		req     *registerUserRequest
		wantID  string
		wantErr error
	}{
		{
			req: &registerUserRequest{
				"username",
				"password1",
				"b@c",
			},
			wantID:  "",
			wantErr: ErrExistingUsernameOrEmail,
		},
		{
			req: &registerUserRequest{
				"username2",
				"password1",
				"a@b",
			},
			wantID:  "",
			wantErr: ErrExistingUsernameOrEmail,
		},
	}

	for _, tt := range tests {
		_, err := suite.svc.RegisterNewUser(suite.req)
		userID, err := suite.svc.RegisterNewUser(*tt.req)

		assert.Equal(suite.T(), tt.wantID, string(userID))
		assert.Equal(suite.T(), tt.wantErr, err)
	}
}

func (suite *ServiceTestSuite) TestRegisterNewUser_AssignsUserANewID() {
	userID, _ := suite.svc.RegisterNewUser(suite.req)

	assert.Greater(suite.T(), len(userID), 2)
}

func (suite *ServiceTestSuite) TestRegisterNewUser_AssignsUserAHashedPassword() {
	userID, _ := suite.svc.RegisterNewUser(suite.req)

	user, err := suite.svc.users.FindByID(userID)

	assert.Nil(suite.T(), err)
	assert.True(suite.T(), checkPasswordHash(user.password, "password"))
}

func (suite *ServiceTestSuite) TestPasswordLength_MustBeAtLeastEight() {
	_, err := suite.svc.RegisterNewUser(registerUserRequest{"username2", "passwod", "b@c"})

	assert.Error(suite.T(), err)
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
