package gomicroblog

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestRegisterNewUser(t *testing.T) {
	Convey("Given new user with username, email and password", t, func() {
		svc := service{users: NewUserRepository()}
		req := registerUserRequest{"username", "password", "user@email.com"}

		Convey("When user registers", func() {
			userID, err := svc.RegisterNewUser(req)

			var created bool
			if err != nil {
				created = false
			} else {
				created = true
			}

			So(created, ShouldEqual, true)
			Convey("Then the created user has username", func() {
				dbUser, err := svc.users.FindByName("username")

				So(err, ShouldBeNil)
				So(userID, ShouldEqual, dbUser.ID)
			})
		})

	})
}

func TestLoginUser(t *testing.T) {
	var req validateUserRequest
	Convey("Given an existing user", t, func() {
		svc := service{users: NewUserRepository()}
		regReq := registerUserRequest{"user", "password", "user@app.com"}
		_, err := svc.RegisterNewUser(regReq)
		So(err, ShouldBeNil)

		Convey("When user provides correct credentials", func() {
			req.Username = "user"
			req.Password = "password"
		})

		Convey("And user does validation", func() {
			userID, err := svc.ValidateUser(req)
			So(err, ShouldBeNil)
			So(IsValidID(string(userID)), ShouldEqual, true)

			Convey("Then the user is successfully validated", func() {
				dbUser, err := svc.users.FindByName(req.Username)
				So(err, ShouldBeNil)
				So(userID, ShouldEqual, dbUser.ID)
			})
		})
	})
}
