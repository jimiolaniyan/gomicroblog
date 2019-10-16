package gomicroblog

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
)

type BddTestSuite struct {
	suite.Suite
	svc    service
	req    registerUserRequest
	userID ID
}

func (bs *BddTestSuite) SetupSuite() {
	bs.svc = service{users: NewUserRepository(), posts: NewPostRepository()}
	bs.req = registerUserRequest{"U", "password", "user@app.com"}

	id, _ := bs.svc.RegisterNewUser(bs.req)
	bs.userID = id
}

func (bs *BddTestSuite) TestRegisterNewUser(t *testing.T) {
	Convey("Given new user with username, email and password", t, func() {

		Convey("When user registers", func() {
			userID, err := bs.svc.RegisterNewUser(bs.req)

			var created bool
			if err != nil {
				created = false
			} else {
				created = true
			}

			So(created, ShouldEqual, true)

			Convey("Then the created user has username", func() {
				dbUser, err := bs.svc.users.FindByName(bs.req.Username)

				So(err, ShouldBeNil)
				So(userID, ShouldEqual, dbUser.ID)
			})
		})

	})
}

func (bs *BddTestSuite) TestLoginUser() {
	var req validateUserRequest
	Convey("Given an existing U", bs.T(), func() {

		Convey("When U provides correct credentials", func() {
			req.Username = bs.req.Username
			req.Password = bs.req.Password

			Convey("And U does validation", func() {
				userID, err := bs.svc.ValidateUser(req)
				So(err, ShouldBeNil)
				So(IsValidID(string(userID)), ShouldEqual, true)

				Convey("Then the U is successfully validated", func() {
					dbUser, err := bs.svc.users.FindByName(req.Username)
					So(err, ShouldBeNil)
					So(userID, ShouldEqual, dbUser.ID)
				})
			})
		})
	})
}

func (bs *BddTestSuite) TestPostCreation() {
	Convey("Given a registered user U with a new post P", bs.T(), func() {
		body := "P"

		Convey("When U creates P", func() {
			postId, err := bs.svc.CreatePost(bs.userID, body)
			So(err, ShouldBeNil)
			So(IsValidID(string(postId)), ShouldBeTrue)

			Convey("Then the user's posts will contain P", func() {
				posts, _ := bs.svc.GetUserPosts(bs.req.Username)
				p := &post{}

				for _, post := range posts {
					if post.Author.Username == "U" && post.body == body {
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

func (bs *BddTestSuite) TestProfileWithNoPosts() {
	Convey("Given a newly registered user U with no posts", bs.T(), func() {
		now := time.Now().UTC()

		Convey("When his profile is requested", func() {
			profile, err := bs.svc.GetProfile(bs.req.Username)
			So(err, ShouldBeNil)
			So(profile, ShouldNotBeNil)

			Convey("Then his profile is as follows", func() {
				expectedProfile := profileResponse{
					Username: "U",
					Avatar:   avatar(bs.req.Email),
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

func (bs *BddTestSuite) TestProfileWithPosts() {
	Convey("Given a returning user U with posts", bs.T(), func() {
		postIDs, ok := createPosts(bs.userID, bs.svc)
		So(ok, ShouldBeTrue)

		err := bs.svc.UpdateLastSeen(bs.userID)
		So(err, ShouldBeNil)

		Convey("When his profile is requested", func() {
			profile, err := bs.svc.GetProfile(bs.req.Username)

			So(err, ShouldBeNil)
			So(profile, ShouldNotBeNil)

			Convey("Then his profile contains his posts in reverse chronological order", func() {

				av := avatar("user@app.com")
				expected := profileResponse{
					Posts: []*post{
						{ID: postIDs[2], Author: Author{Username: "U", UserID: bs.userID, Avatar: av}, body: "C", timestamp: profile.Posts[0].timestamp},
						{ID: postIDs[1], Author: Author{Username: "U", UserID: bs.userID, Avatar: av}, body: "B", timestamp: profile.Posts[1].timestamp},
						{ID: postIDs[0], Author: Author{Username: "U", UserID: bs.userID, Avatar: av}, body: "A", timestamp: profile.Posts[2].timestamp},
					},
				}

				So(err, ShouldBeNil)
				So(expected.Posts, ShouldResemble, profile.Posts)

				Convey("Add his last seen is updated.", func() {
					user, _ := bs.svc.users.FindByID(bs.userID)
					So(profile.LastSeen, ShouldEqual, user.lastSeen)
					So(profile.LastSeen.After(profile.Joined), ShouldBeTrue)
				})
			})
		})
	})
}

func (bs *BddTestSuite) TestEditUserProfile() {
	Convey("Given a returning user U", bs.T(), func() {

		Convey("When the user edits his profile", func() {
			bio := "My wonderful bio"
			err := bs.svc.EditProfile(bs.userID, editProfileRequest{Username: "U2", Bio: bio})

			So(err, ShouldBeNil)

			Convey("Then his profile shows the updated information", func() {
				user, _ := bs.svc.users.FindByID(bs.userID)
				So(user.username, ShouldEqual, "U2")
				So(user.bio, ShouldEqual, bio)
			})
		})
	})
}

func createPosts(id ID, svc service) (ids []PostID, ok bool) {
	id1, _ := svc.CreatePost(id, "A")
	id2, _ := svc.CreatePost(id, "B")
	id3, _ := svc.CreatePost(id, "C")

	ids = append(ids, id1, id2, id3)
	ok = true
	return
}
