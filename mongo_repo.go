package gomicroblog

import (
	"context"
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
	Followers map[ID]*user
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
	panic("implement me")
}

func (m *mongoUserRepository) FindFriends(username string) ([]user, error) {
	panic("implement me")
}

func (m *mongoUserRepository) FindFollowers(username string) ([]user, error) {
	panic("implement me")
}

func dbUserFromUser(u *user) dbUser {
	var ids []ID
	for _, id := range u.Friends {
		ids = append(ids, id)
	}
	//return dbUser{u.ID, u.username, u.password, u.email, u.createdAt, u.lastSeen, u.bio, ids, u.Followers}
	return dbUser{}
}

func userFromDBUser(u dbUser) user {
	friends := map[ID]*user{}
	for _, id := range u.Friends {
		friends[id] = &user{}
	}
	//nU := user{u.ID, u.Username, u.Password, u.Email, u.CreatedAt, u.LastSeen, u.Bio, friends, u.Followers}
	return user{}
}
