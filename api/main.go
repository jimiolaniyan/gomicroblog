package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/julienschmidt/httprouter"

	. "github.com/jimiolaniyan/gomicroblog"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	u := client.Database("microblog").Collection("users")
	p := client.Database("microblog").Collection("posts")

	svc := NewService(NewMongoUserRepository(u), NewMongoPostRepository(p))

	router := httprouter.New()
	router.Handler(http.MethodPost, "/v1/users", RegisterUserHandler(svc))
	router.Handler(http.MethodPost, "/v1/sessions", LoginHandler(svc))
	router.Handler(http.MethodPost, "/v1/posts", RequireAuth(LastSeenMiddleware(CreatePostHandler(svc), svc)))
	router.Handler(http.MethodGet, "/v1/users/:username", LastSeenMiddleware(GetProfileHandler(svc), svc))
	router.Handler(http.MethodPatch, "/v1/users", RequireAuth(LastSeenMiddleware(EditProfileHandler(svc), svc)))
	router.Handler(http.MethodGet, "/v1/users/:username/followers", RequireAuth(LastSeenMiddleware(GetUserFollowersHandler(svc), svc)))
	router.Handler(http.MethodGet, "/v1/users/:username/friends", RequireAuth(LastSeenMiddleware(GetUserFriendsHandler(svc), svc)))
	router.Handler(http.MethodPost, "/v1/users/:username/followers", RequireAuth(LastSeenMiddleware(CreateRelationshipHandler(svc), svc)))
	router.Handler(http.MethodDelete, "/v1/users/:username/followers", RequireAuth(LastSeenMiddleware(RemoveRelationshipHandler(svc), svc)))

	log.Printf("Server started. Listening on port: %s\n", "8090")
	log.Fatal(http.ListenAndServe(":"+("8090"), router))
}
