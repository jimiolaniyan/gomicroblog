package gomicroblog

type userRepository struct {
	users map[ID]*user
}

func (repo *userRepository) FindByID(id ID) (*user, error) {
	if u, ok := repo.users[id]; ok {
		return u, nil
	}
	return nil, ErrNotFound
}

func (repo *userRepository) FindByEmail(email string) (*user, error) {
	for _, v := range repo.users {
		if v.email == email {
			return v, nil
		}
	}
	return nil, ErrNotFound
}

func (repo *userRepository) Store(user *user) error {
	repo.users[user.ID] = user
	return nil
}

func (repo *userRepository) FindByName(username string) (*user, error) {
	for _, v := range repo.users {
		if v.username == username {
			return v, nil
		}
	}
	return nil, ErrNotFound
}

func NewUserRepository() Repository {
	return &userRepository{users: map[ID]*user{}}
}
