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

type postRepository struct {
	posts map[PostID]*post
}

func (repo *postRepository) Store(post *post) error {
	repo.posts[post.ID] = post
	return nil
}

func (repo *postRepository) FindByID(id PostID) (*post, error) {
	if p, ok := repo.posts[id]; ok {
		return p, nil
	}
	return nil, ErrPostNotFound
}

func (repo *postRepository) FindByUserID(id ID) ([]*post, error) {
	var posts []*post
	for _, p := range repo.posts {
		if p.Author.UserID == id {
			posts = append(posts, p)
		}
	}
	return posts, nil
}

func NewPostRepository() PostRepository {
	return &postRepository{posts: map[PostID]*post{}}
}
