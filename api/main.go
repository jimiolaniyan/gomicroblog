package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jimiolaniyan/gomicroblog/auth"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	. "github.com/jimiolaniyan/gomicroblog"
)

var dbURL = os.Getenv("DATABASE_URL")
var dbName = os.Getenv("DATABASE_NAME")

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+dbURL))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	u := client.Database(dbName).Collection("users")
	p := client.Database(dbName).Collection("posts")

	svc := NewService(NewMongoUserRepository(u), NewMongoPostRepository(p))
	authSvc := auth.NewService(auth.NewAccountRepository(), NewAccountCreatedHandler(svc))

	router := httprouter.New()
	router.Handler(http.MethodPost, "/auth/v1/accounts", auth.RegisterAccountHandler(authSvc))
	router.Handler(http.MethodPost, "/auth/v1/sessions", auth.LoginHandler(authSvc))
	router.Handler(http.MethodPost, "/v1/posts", RequireAuth(LastSeenMiddleware(CreatePostHandler(svc), svc)))
	router.Handler(http.MethodGet, "/v1/users/:username", LastSeenMiddleware(GetProfileHandler(svc), svc))
	router.Handler(http.MethodPatch, "/v1/users", RequireAuth(LastSeenMiddleware(EditProfileHandler(svc), svc)))
	router.Handler(http.MethodGet, "/v1/users/:username/followers", RequireAuth(LastSeenMiddleware(GetUserFollowersHandler(svc), svc)))
	router.Handler(http.MethodGet, "/v1/users/:username/friends", RequireAuth(LastSeenMiddleware(GetUserFriendsHandler(svc), svc)))
	router.Handler(http.MethodPost, "/v1/users/:username/followers", RequireAuth(LastSeenMiddleware(CreateRelationshipHandler(svc), svc)))
	router.Handler(http.MethodDelete, "/v1/users/:username/followers", RequireAuth(LastSeenMiddleware(RemoveRelationshipHandler(svc), svc)))

	log.Printf("Server started. Listening on port: %s\n", "8090")
	log.Fatal(http.ListenAndServe(":"+"8090", router))
}
