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
	router.Handler(http.MethodPost, "/v1/users", RegisterUserHandler(svc))
	router.Handler(http.MethodPost, "/v1/sessions", LoginHandler(svc))
	router.Handler(http.MethodPost, "/v1/posts", RequireAuth(LastSeenMiddleware(CreatePostHandler(svc), svc)))
	router.Handler(http.MethodGet, "/v1/users/:username", LastSeenMiddleware(GetProfileHandler(svc), svc))
	router.Handler(http.MethodPatch, "/v1/users", RequireAuth(LastSeenMiddleware(EditProfileHandler(svc), svc)))
	router.Handler(http.MethodPost, "/v1/users/:username/followers", RequireAuth(LastSeenMiddleware(CreateRelationshipHandler(svc), svc)))
	router.Handler(http.MethodDelete, "/v1/users/:username/followers", RequireAuth(LastSeenMiddleware(RemoveRelationshipHandler(svc), svc)))

	log.Printf("Server started. Listening on port: %s\n", "8090")
	log.Fatal(http.ListenAndServe(":"+("8090"), router))
}
