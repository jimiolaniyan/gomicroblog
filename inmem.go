package gomicroblog

import (
	"sort"
)

type userRepository struct {
	users map[ID]*User
}

func NewUserRepository() Repository {
	return &userRepository{users: map[ID]*User{}}
}

func (repo *userRepository) FindByID(id ID) (*User, error) {
	if u, ok := repo.users[id]; ok {
		return u, nil
	}
	return nil, ErrNotFound
}

func (repo *userRepository) FindByEmail(email string) (*User, error) {
	for _, v := range repo.users {
		if v.email == email {
			return v, nil
		}
	}
	return nil, ErrNotFound
}

func (repo *userRepository) Store(user *User) error {
	repo.users[user.ID] = user
	return nil
}

func (repo *userRepository) FindByName(username string) (*User, error) {
	for _, v := range repo.users {
		if v.username == username {
			return v, nil
		}
	}
	return nil, ErrNotFound
}

func (repo *userRepository) Update(u *User) error {
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

func (repo *userRepository) FindByIDs(ids []ID) ([]User, error) {
	users := []User{}
	for _, id := range ids {
		if u, _ := repo.FindByID(id); u != nil {
			users = append(users, *u)
		}

	}
	return users, nil
}

type postRepository struct {
	posts map[PostID]Post
}

func NewPostRepository() PostRepository {
	return &postRepository{posts: map[PostID]Post{}}
}

func (repo *postRepository) Store(post Post) error {
	repo.posts[post.ID] = post
	return nil
}

func (repo *postRepository) FindByID(id PostID) (Post, error) {
	if p, ok := repo.posts[id]; ok {
		return p, nil
	}
	return Post{}, ErrPostNotFound
}

func (repo *postRepository) FindLatestPostsForUser(id ID) ([]*Post, error) {
	posts := repo.FindUserPosts(id)

	sortPostsByTimestamp(posts)

	return posts, nil
}

func (repo *postRepository) FindLatestPostsForUserAndFriends(user *User) ([]*Post, error) {
	id := user.ID
	posts := repo.FindUserPosts(id)

	for _, id := range user.Friends {
		ps := repo.FindUserPosts(id)
		posts = append(posts, ps...)
	}

	sortPostsByTimestamp(posts)
	return posts, nil
}

func (repo *postRepository) FindUserPosts(id ID) []*Post {
	var posts []*Post
	for i, p := range repo.posts {
		if p.Author.UserID == id {
			pp := repo.posts[i]
			posts = append(posts, &pp)
		}
	}
	return posts
}

func sortPostsByTimestamp(posts []*Post) {
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Timestamp.After(posts[j].Timestamp)
	})
}
