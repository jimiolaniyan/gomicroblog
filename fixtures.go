package gomicroblog

func duplicateUser(s service, u user, username string) *user {
	u1 := u
	u1.ID = nextID()
	u1.username = username

	u1.Friends = map[ID]*user{}
	u1.Followers = map[ID]*user{}
	_ = s.users.Store(&u1)

	return &u1
}
