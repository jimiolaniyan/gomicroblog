package gomicroblog

type userRepository struct {
	users map[string]*user
}

func (repo *userRepository) FindByEmail(e string) (*user, error) {
	for _, v := range repo.users {
		if v.email == e {
			return v, nil
		}
	}
	return nil, ErrNotFound
}

func (repo *userRepository) Store(user *user) error {
	repo.users[user.username] = user
	return nil
}

func (repo *userRepository) FindByName(username string) (*user, error) {
	if u, ok := repo.users[username]; ok {
		return u, nil
	}
	return nil, ErrNotFound
}

func NewUserRepository() Repository {
	return &userRepository{users: map[string]*user{}}
}
