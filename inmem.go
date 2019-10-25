package gomicroblog

import (
	"sort"
)

type userRepository struct {
	users map[ID]*user
}

func NewUserRepository() Repository {
	return &userRepository{users: map[ID]*user{}}
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

func (repo *userRepository) Update(u *user) error {
	// We don't need to do anything for in-memory implementations
	// since updating is taken care of when using pointers
	return nil
}

func (repo *userRepository) Delete(id ID) error {
	if _, ok := repo.users[id]; !ok {
		return ErrNotFound
	}
	delete(repo.users, id)
	return nil
}

func (repo *userRepository) FindByIDs(ids []ID) ([]user, error) {
	users := []user{}
	for _, id := range ids {
		if u, _ := repo.FindByID(id); u != nil {
			users = append(users, *u)
		}

	}
	return users, nil
}

type postRepository struct {
	posts map[PostID]post
}

func NewPostRepository() PostRepository {
	return &postRepository{posts: map[PostID]post{}}
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
	posts := repo.FindUserPosts(id)

	sortPostsByTimestamp(posts)

	return posts, nil
}

func (repo *postRepository) FindLatestPostsForUserAndFriends(user *user) ([]*post, error) {
	id := user.ID
	posts := repo.FindUserPosts(id)

	for _, id := range user.Friends {
		ps := repo.FindUserPosts(id)
		posts = append(posts, ps...)
	}

	sortPostsByTimestamp(posts)
	return posts, nil
}

func (repo *postRepository) FindUserPosts(id ID) []*post {
	var posts []*post
	for i, p := range repo.posts {
		if p.Author.UserID == id {
			pp := repo.posts[i]
			posts = append(posts, &pp)
		}
	}
	return posts
}

func sortPostsByTimestamp(posts []*post) {
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].timestamp.After(posts[j].timestamp)
	})
}
