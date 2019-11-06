package auth

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRegisterNewUser(t *testing.T) {

	convey.Convey("Given new user with username, email and password", t, func() {
		req := registerAccountRequest{"user", "user@user.com", "password"}
		accounts := NewAccountRepository()
		svc := NewService(accounts, &accountEventsSpy{})

		convey.Convey("When user registers", func() {
			userID, err := svc.RegisterAccount(req)

			convey.So(err, convey.ShouldBeNil)

			convey.Convey("Then the created user has username", func() {
				dbUser, err := accounts.FindByName(req.Username)

				convey.So(err, convey.ShouldBeNil)
				convey.So(userID, convey.ShouldEqual, dbUser.ID)
			})
		})

	})
}
