package auth

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRegisterNewUser(t *testing.T) {
	convey.Convey("Given new user with username, email and password", t, func() {
		req := registerAccountRequest{"user", "user@user.com", "password"}
		accounts := NewAccountRepository()
		svc := NewService(accounts, &eventsSpy{})

		convey.Convey("When user registers", func() {
			id, err := svc.RegisterAccount(req)

			convey.So(err, convey.ShouldBeNil)

			convey.Convey("Then the created user has username", func() {
				acc, err := accounts.FindByName(req.Username)

				convey.So(err, convey.ShouldBeNil)
				convey.So(id, convey.ShouldEqual, acc.ID)
			})
		})

	})
}

func TestLoginUser(t *testing.T) {
	convey.Convey("Given an existing U", t, func() {
		username := "user"
		accounts := NewAccountRepository()
		svc := NewService(accounts, &eventsSpy{})
		id, err := svc.RegisterAccount(registerAccountRequest{username, "user@user.com", "password"})

		convey.So(err, convey.ShouldBeNil)
		convey.So(isValidID(string(id)), convey.ShouldBeTrue)

		convey.Convey("When U provides correct credentials", func() {
			req := validateCredentialsRequest{username, "password"}

			convey.Convey("And U does validation", func() {
				id, err := svc.ValidateCredentials(req)
				convey.So(err, convey.ShouldBeNil)
				convey.So(isValidID(string(id)), convey.ShouldBeTrue)

				convey.Convey("Then the U is successfully validated", func() {
					acc, err := accounts.FindByName(username)
					convey.So(err, convey.ShouldBeNil)
					convey.So(id, convey.ShouldEqual, acc.ID)
				})
			})
		})
	})
}
