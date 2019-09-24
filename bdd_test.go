package gomicroblog

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
	"testing"
)

type UserTestSuite struct {
	suite.Suite
	svc service
}

func (suite *UserTestSuite) SetupTest() {
	suite.svc = service{users: NewUserRepository()}
}

func (suite *UserTestSuite) TestRegisterNewUser() {
	Convey("Given new user with username, email and password", suite.T(), func() {
		username := "username"
		email := "user@email.com"
		password := "password"

		Convey("When user registers", func() {
			user, err := suite.svc.RegisterNewUser(username, password, email)

			var created bool
			if err != nil {
				created = false
			} else {
				created = true
			}

			So(created, ShouldEqual, true)
			Convey("Then the created user has username", func() {
				dbUser, err := suite.svc.users.FindByName(username)

				So(err, ShouldBeNil)
				So(user.username, ShouldEqual, dbUser.username)
			})
		})

	})
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
