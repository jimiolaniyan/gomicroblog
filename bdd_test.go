package gomicroblog

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewUserAccountCreation(t *testing.T) {
	Convey("Given new user with username, email and password", t, func() {
		username := "username"
		email := "user@email.com"
		password := "password"
		Convey("When user creates account", func() {
			user, err := Register(username, email,password);
			var created bool
			if err != nil {
				created = false
			} else {
				created = true
			}

			So(created, ShouldEqual, true)
			Convey("Then the created user has username", func() {
				So(user.username, ShouldEqual, username)
			})
		})

	})
}
