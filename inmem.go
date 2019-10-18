package gomicroblog

import (
	"sort"
)

type userRepository struct {
	users map[ID]*user
}

func (repo *userRepository) Delete(id ID) error {
	if _, ok := repo.users[id]; !ok {
		return ErrNotFound
	}
	delete(repo.users, id)
	return nil
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
	posts map[PostID]post
}

func (repo *postRepository) Store(post post) error {
	repo.posts[post.ID] = post
	return nil
}

func (repo *postRepository) FindByID(id PostID) (post, error) {
	if p, ok := repo.posts[id]; ok {
		return p, nil
	}
	return post{}, ErrPostNotFound
}

func (repo *postRepository) FindLatestPostsForUser(id ID) ([]*post, error) {
	//fmt.Println(id)
	var posts []*post
	for i, p := range repo.posts {
		if p.Author.UserID == id {
			pp := repo.posts[i]
			posts = append(posts, &pp)
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].timestamp.After(posts[j].timestamp)
	})

	return posts, nil
}

func NewPostRepository() PostRepository {
	return &postRepository{posts: map[PostID]post{}}
}
