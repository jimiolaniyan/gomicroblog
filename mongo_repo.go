package gomicroblog

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
)

type mongoUserRepository struct {
	collection *mongo.Collection
}

type dbUser struct {
	ID        ID `bson:"_id"`
	Username  string
	Password  string
	Email     string
	CreatedAt time.Time
	LastSeen  time.Time
	Bio       string
	Friends   []ID
	Followers []ID
}

func NewMongoUserRepository(c *mongo.Collection) Repository {
	return &mongoUserRepository{collection: c}
}

func (m *mongoUserRepository) FindByName(username string) (*user, error) {
	return m.findUserBy("username", username)
}

func (m *mongoUserRepository) FindByEmail(email string) (*user, error) {
	return m.findUserBy("email", email)
}

func (m *mongoUserRepository) FindByID(id ID) (*user, error) {
	return m.findUserBy("_id", string(id))
}

func (m *mongoUserRepository) Update(u *user) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbu := dbUserFromUser(u)
	_, err := m.collection.ReplaceOne(ctx, bson.M{"_id": dbu.ID}, dbu)
	return err
}

func (m *mongoUserRepository) Store(u *user) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbu := dbUserFromUser(u)
	_, err := m.collection.InsertOne(ctx, &dbu)
	return err
}

func (m *mongoUserRepository) Delete(id ID) error {
	return nil
}

func (m *mongoUserRepository) FindByIDs(ids []ID) ([]user, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	friends := []user{}

	filter := bson.D{{"_id", bson.D{
		{"$in", ids},
	}}}

	cursor, err := m.collection.Find(ctx, filter)
	if err != nil {
		return friends, err
	}

	for cursor.Next(ctx) {
		var u dbUser
		err := cursor.Decode(&u)
		if err != nil {
			return friends, err
		}

		friends = append(friends, userFromDBUser(u))
	}
	return friends, nil
}

func (m *mongoUserRepository) findUserBy(key string, val string) (*user, error) {
	var u dbUser

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sr := m.collection.FindOne(ctx, bson.M{key: val})

	if sr.Err() == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}

	if err := sr.Decode(&u); err != nil {
		return nil, err
	}

	nU := userFromDBUser(u)
	return &nU, nil
}

func dbUserFromUser(u *user) dbUser {
	return dbUser{u.ID, u.username, u.password, u.email, u.createdAt, u.lastSeen, u.bio, u.Friends, u.Followers}
}

func userFromDBUser(u dbUser) user {
	return user{u.ID, u.Username, u.Password, u.Email, u.CreatedAt, u.LastSeen, u.Bio, u.Friends, u.Followers}
}

type mongoPostRepository struct {
	collection *mongo.Collection
}

func NewMongoPostRepository(c *mongo.Collection) PostRepository {
	return &mongoPostRepository{collection: c}
}

func (m *mongoPostRepository) FindByID(id PostID) (post, error) {
	return post{}, errors.New("implement me")
}

func (m *mongoPostRepository) Store(post post) error {
	_, err := m.collection.InsertOne(context.TODO(), &post)
	return err
}

func (m *mongoPostRepository) FindLatestPostsForUser(id ID) ([]*post, error) {
	filter := bson.D{
		{"author.user_id", id},
	}

	cursor, err := m.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	posts := []*post{}
	for cursor.Next(context.TODO()) {
		var p post
		err := cursor.Decode(&p)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &p)
	}

	return posts, nil
}

func (m *mongoPostRepository) FindLatestPostsForUserAndFriends(user *user) ([]*post, error) {
	return nil, errors.New("implement me")
}
