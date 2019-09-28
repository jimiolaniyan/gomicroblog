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

func (suite *ServiceTestSuite) TestExistingUsername_CannotBeReused() {
	_, err := suite.svc.RegisterNewUser(suite.req)
	user, err := suite.svc.RegisterNewUser(registerUserRequest{"username", "password1", "b@c"})

	assert.Nil(suite.T(), user)
	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestExistingEmail_CannotBeReused() {
	_, err := suite.svc.RegisterNewUser(suite.req)
	user2, err := suite.svc.RegisterNewUser(registerUserRequest{"username2", "password1", "a@b"})

	assert.Nil(suite.T(), user2)
	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestRegisterNewUser_AssignsUserANewID() {
	user1, _ := suite.svc.RegisterNewUser(suite.req)

	assert.Greater(suite.T(), len(user1.ID), 2)
}

func (suite *ServiceTestSuite) TestRegisterNewUser_AssignsUserAHashedPassword() {
	user1, _ := suite.svc.RegisterNewUser(suite.req)

	assert.True(suite.T(), checkPasswordHash(user1.password, "password"))
}

func (suite *ServiceTestSuite) TestPasswordLength_MustBeAtLeastEight() {
	_, err := suite.svc.RegisterNewUser(registerUserRequest{"username2", "passwod", "b@c"})

	assert.Error(suite.T(), err)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
