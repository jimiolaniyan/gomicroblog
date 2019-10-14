package gomicroblog

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
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

func TestPostCreation(t *testing.T) {
	Convey("Given a registered user U", t, func() {
		svc := NewService(NewUserRepository(), NewPostRepository())
		userID, err := svc.RegisterNewUser(registerUserRequest{"U", "password", "user@app.com"})
		So(err, ShouldBeNil)
		So(IsValidID(string(userID)), ShouldBeTrue)

		Convey("With a new post P", nil)
		body := "P"

		Convey("When U creates P", func() {
			postId, err := svc.CreatePost(userID, body)
			So(err, ShouldBeNil)
			So(IsValidID(string(postId)), ShouldBeTrue)

			Convey("Then the user's posts will contain P", func() {
				posts, _ := svc.GetUserPosts("U")
				p := &post{}

				for _, post := range posts {
					if post.Author.Username == "U" {
						p = post
					}
				}

				So(p, ShouldNotBeNil)
				So(postId, ShouldEqual, p.ID)
				So(body, ShouldEqual, p.body)
			})
		})
	})
}

func TestProfileWithNoPosts(t *testing.T) {
	Convey("Given a newly registered user U with no posts", t, func() {
		now := time.Now().UTC()
		svc := NewService(NewUserRepository(), NewPostRepository())
		email := "user@app.com"
		userID, err := svc.RegisterNewUser(registerUserRequest{"U", "password", email})
		So(err, ShouldBeNil)
		So(IsValidID(string(userID)), ShouldBeTrue)

		Convey("When his profile is requested", func() {
			profile, err := svc.GetProfile("U")
			So(err, ShouldBeNil)
			So(profile, ShouldNotBeNil)

			Convey("Then his profile is as follows", func() {
				expectedProfile := profileResponse{
					Username: "U",
					Avatar:   avatar(email),
					Bio:      "",
					Joined:   profile.Joined,
					LastSeen: profile.LastSeen,
				}

				So(profile, ShouldResemble, expectedProfile)
				So(profile.Joined.After(now), ShouldBeTrue)
				So(profile.LastSeen.After(now), ShouldBeTrue)
			})

		})

	})
}
