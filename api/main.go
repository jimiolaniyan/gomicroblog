package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	. "github.com/jimiolaniyan/gomicroblog"
)

func main() {
	svc := NewService(NewUserRepository(), NewPostRepository())

	router := httprouter.New()
	router.Handler(http.MethodPost, "/v1/users/new", RegisterUserHandler(svc))
	router.Handler(http.MethodPost, "/v1/auth/login", LoginHandler(svc))
	router.Handler(http.MethodPost, "/v1/posts", RequireAuth(LastSeenMiddleware(CreatePostHandler(svc), svc)))
	router.Handler(http.MethodGet, "/v1/users/:username", RequireAuth(LastSeenMiddleware(GetProfileHandler(svc), svc)))

	log.Printf("Server started. Listening on port: %s\n", "8090")
	log.Fatal(http.ListenAndServe(":"+("8090"), router))
}
