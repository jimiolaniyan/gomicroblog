package blog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
)

type BddTestSuite struct {
	suite.Suite
	svc                       service
	username, password, email string
	now                       time.Time
	userID                    ID
	user                      *User
}

func (bs *BddTestSuite) SetupSuite() {
	bs.now = time.Now().UTC()
	bs.svc = service{users: NewUserRepository(), posts: NewPostRepository()}

	bs.userID = nextID()
	bs.username = "U"
	bs.email = "user@app.com"

	bs.svc.CreateProfile(string(bs.userID), bs.username, bs.email)

	u, _ := bs.svc.users.FindByID(bs.userID)
	bs.password = "password"
	u.Password, _ = hashPassword(bs.password)
	bs.user = u
}

func (bs *BddTestSuite) TearDownTest() {
	bs.svc.posts = NewPostRepository()
}

func (bs *BddTestSuite) TestPostCreation() {
	Convey("Given a registered user U with a new post P", bs.T(), func() {
		body := "P"

		Convey("When U creates P", func() {
			postID, err := bs.svc.CreatePost(bs.userID, body)
			So(err, ShouldBeNil)
			So(IsValidID(string(postID)), ShouldBeTrue)

			Convey("Then the user's posts will contain P", func() {
				posts, _ := bs.svc.GetUserPosts(bs.username)
				p := &Post{}

				for _, post := range posts {
					if post.Author.UserID == bs.userID && post.Body == body {
						p = post
					}
				}

				So(p, ShouldNotBeNil)
				So(postID, ShouldEqual, p.ID)
				So(body, ShouldEqual, p.Body)
			})
		})
	})
}

func (bs *BddTestSuite) TestProfileWithNoPosts() {
	Convey("Given a newly registered user U with no posts", bs.T(), func() {

		Convey("When his profile is requested", func() {
			profile, err := bs.svc.GetProfile(bs.username)
			So(err, ShouldBeNil)
			So(profile, ShouldNotBeNil)

			Convey("Then his profile is as follows", func() {
				expectedProfile := Profile{
					ID:       bs.userID,
					Username: "U",
					Avatar:   avatar(bs.email),
					Bio:      "",
					Joined:   profile.Joined,
					LastSeen: profile.LastSeen,
					Posts:    []postResponse{},
				}

				So(profile, ShouldResemble, expectedProfile)
				So(profile.Joined.After(bs.now), ShouldBeTrue)
				So(profile.LastSeen.After(bs.now), ShouldBeTrue)
			})

		})

	})
}

func (bs *BddTestSuite) TestProfileWithPostsAndFriends() {
	Convey("Given a returning user U1 with posts", bs.T(), func() {
		u1 := DuplicateUser(bs.svc.users, *bs.user, "fU1")

		postIDs, ok := createPosts(u1.ID, bs.svc)
		So(ok, ShouldBeTrue)

		err := bs.svc.UpdateLastSeen(u1.ID)
		So(err, ShouldBeNil)

		Convey("With U1 and U2 following each other", func() {
			u2 := DuplicateUser(bs.svc.users, *bs.user, "fU2")
			u1.Follow(u2)
			u2.Follow(u1)
			Convey("When his profile is requested", func() {
				profile, err := bs.svc.GetProfile(u1.Username)

				So(err, ShouldBeNil)
				So(profile, ShouldNotBeNil)

				Convey("Then his profile contains his posts in reverse chronological order", func() {
					ar := authorResponse{Username: u1.Username, UserID: u1.ID, Avatar: avatar("user@app.com")}
					expected := Profile{
						Relationships: Relationships{Followers: 1, Friends: 1},
						Posts: []postResponse{
							{ID: postIDs[2], Author: ar, Body: "C", Timestamp: profile.Posts[0].Timestamp},
							{ID: postIDs[1], Author: ar, Body: "B", Timestamp: profile.Posts[1].Timestamp},
							{ID: postIDs[0], Author: ar, Body: "A", Timestamp: profile.Posts[2].Timestamp},
						},
					}

					So(err, ShouldBeNil)
					So(expected.Posts, ShouldResemble, profile.Posts)

					Convey("Add his last seen is updated.", func() {
						user, _ := bs.svc.users.FindByID(u1.ID)
						So(profile.LastSeen, ShouldEqual, user.LastSeen)
						So(profile.LastSeen.After(profile.Joined), ShouldBeTrue)

						Reset(func() {
							_ = bs.svc.users.Delete(u1.ID)
							_ = bs.svc.users.Delete(u2.ID)
						})
					})
				})
			})
		})
	})
}

func (bs *BddTestSuite) TestEditUserProfile() {
	Convey("Given a returning user U", bs.T(), func() {
		existingUser := *bs.user
		existingUser.ID = nextID()
		existingUser.Username = "newUsername"

		err := bs.svc.users.Store(&existingUser)
		assert.Nil(bs.T(), err)

		Convey("When the user edits his profile", func() {
			bio := "My wonderful bio"
			newU := "U2"
			err := bs.svc.EditProfile(existingUser.ID, editProfileRequest{Username: &newU, Bio: &bio})

			So(err, ShouldBeNil)

			Convey("Then his profile shows the updated information", func() {
				profile, err := bs.svc.GetProfile(existingUser.Username)

				So(err, ShouldBeNil)
				So(profile.Username, ShouldEqual, newU)
				So(profile.Bio, ShouldEqual, bio)

				Reset(func() {
					_ = bs.svc.users.Delete(existingUser.ID)
				})
			})
		})
	})
}

func (bs *BddTestSuite) TestRelationships_Create() {
	Convey("Given two users U1 and U2 with no relationship", bs.T(), func() {
		u1 := DuplicateUser(bs.svc.users, *bs.user, "U1")
		u2 := DuplicateUser(bs.svc.users, *bs.user, "U2")

		Convey("When U1 follows U2", func() {

			err := bs.svc.CreateRelationshipFor(u1.ID, u2.Username)
			So(err, ShouldBeNil)

			Convey("Then U1 is following U2", func() {
				So(u1.IsFollowing(u2), ShouldBeTrue)

				Convey("And U2 is in U1's friends list", func() {
					friends, err := bs.svc.GetUserFriends(u1.Username)

					So(err, ShouldBeNil)

					userInfo := getUserInfoFromList(friends, u2.ID)
					expectedUserInfo := createInfoFromUser(u2)

					So(userInfo, ShouldResemble, expectedUserInfo)

					Convey("And U1 is in U2's followers list", func() {
						followers, err := bs.svc.GetUserFollowers(u2.Username)

						So(err, ShouldBeNil)

						userInfo := getUserInfoFromList(followers, u1.ID)
						expectedUserInfo := createInfoFromUser(u1)

						So(userInfo, ShouldResemble, expectedUserInfo)

						Reset(func() {
							_ = bs.svc.users.Delete(u2.ID)
							_ = bs.svc.users.Delete(u1.ID)
						})
					})
				})
			})
		})
	})
}

func (bs *BddTestSuite) TestRelationships_Remove() {
	Convey("Given two users U1 and U1", bs.T(), func() {
		u1 := DuplicateUser(bs.svc.users, *bs.user, "newU1")
		u2 := DuplicateUser(bs.svc.users, *bs.user, "newU2")

		Convey("With U1 following U2", func() {
			u1.Follow(u2)
			So(u1.IsFollowing(u2), ShouldBeTrue)

			Convey("When U1 unfollows u2", func() {
				err := bs.svc.RemoveRelationshipFor(u1.ID, u2.Username)
				So(err, ShouldBeNil)

				Convey("Then U1 is not following U2", func() {
					So(u1.IsFollowing(u2), ShouldBeFalse)

					Convey("And U2 is not in U1's friends list", func() {
						friends, err := bs.svc.GetUserFriends(u1.Username)

						So(err, ShouldBeNil)

						userInfo := getUserInfoFromList(friends, u2.ID)
						So(userInfo, ShouldResemble, UserInfo{})

						Convey("And U1 is not in U2's follower's list", func() {
							followers, err := bs.svc.GetUserFollowers(u2.Username)

							So(err, ShouldBeNil)

							userInfo := getUserInfoFromList(followers, u1.ID)
							So(userInfo, ShouldResemble, UserInfo{})

							Reset(func() {
								_ = bs.svc.users.Delete(u1.ID)
								_ = bs.svc.users.Delete(u2.ID)
							})
						})
					})
				})
			})
		})
	})
}

func (bs *BddTestSuite) TestTimelines() {
	Convey("Given user U1 following U2 and U3 ", bs.T(), func() {
		u1 := DuplicateUser(bs.svc.users, *bs.user, "uu1")
		u2 := DuplicateUser(bs.svc.users, *bs.user, "uu2")
		u3 := DuplicateUser(bs.svc.users, *bs.user, "uu3")

		u1.Follow(u2)
		u1.Follow(u3)
		Convey("With U1, U2 and U3 having the following posts", func() {
			posts := []string{"p1", "p2", "p3", "p4", "p5", "p6"}
			p21ID, _ := bs.svc.CreatePost(u2.ID, posts[1])
			p31ID, _ := bs.svc.CreatePost(u3.ID, posts[0])
			p22ID, _ := bs.svc.CreatePost(u2.ID, posts[5])
			p11ID, _ := bs.svc.CreatePost(u1.ID, posts[2])
			p32ID, _ := bs.svc.CreatePost(u3.ID, posts[3])
			p12ID, _ := bs.svc.CreatePost(u2.ID, posts[4])

			Convey("When requests his timeline", func() {
				tl, err := bs.svc.GetTimeline(u1.ID)
				So(err, ShouldBeNil)

				Convey("Then his timeline is as follows", func() {
					ar := authorResponse{u1.ID, u1.Username, avatar(u1.Email)}
					expectedTL := []postResponse{
						{p12ID, posts[4], tl[0].Timestamp, ar},
						{p32ID, posts[3], tl[1].Timestamp, ar},
						{p11ID, posts[2], tl[2].Timestamp, ar},
						{p22ID, posts[5], tl[3].Timestamp, ar},
						{p31ID, posts[0], tl[4].Timestamp, ar},
						{p21ID, posts[1], tl[5].Timestamp, ar},
					}

					So(tl, ShouldResemble, expectedTL)

					Reset(func() {
						_ = bs.svc.users.Delete(u1.ID)
						_ = bs.svc.users.Delete(u2.ID)
						_ = bs.svc.users.Delete(u3.ID)
					})
				})
			})
		})

	})
}

func createInfoFromUser(u2 *User) UserInfo {
	return UserInfo{
		ID:       u2.ID,
		Username: u2.Username,
		Avatar:   avatar(u2.Email),
		Bio:      u2.Bio,
		Joined:   u2.CreatedAt,
	}
}

func getUserInfoFromList(infos []UserInfo, id ID) UserInfo {
	var t UserInfo
	for _, userInfo := range infos {
		if userInfo.ID == id {
			t = userInfo
		}
	}
	return t
}

func createPosts(id ID, svc service) (ids []PostID, ok bool) {
	id1, _ := svc.CreatePost(id, "A")
	id2, _ := svc.CreatePost(id, "B")
	id3, _ := svc.CreatePost(id, "C")

	ids = append(ids, id1, id2, id3)
	ok = true
	return
}

func TestBddSuite(t *testing.T) {
	suite.Run(t, new(BddTestSuite))
}
