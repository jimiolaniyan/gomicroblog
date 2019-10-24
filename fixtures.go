package gomicroblog

func duplicateUser(s service, u user, username string) *user {
	u1 := u
	u1.ID = nextID()
	u1.username = username

	_ = s.users.Store(&u1)

	return &u1
}
