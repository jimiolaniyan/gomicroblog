package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	blog "github.com/jimiolaniyan/gomicroblog"
)

func main() {
	svc := blog.NewService(blog.NewUserRepository(), blog.NewPostRepository())

	router := httprouter.New()
	router.Handler(http.MethodPost, "/v1/users/new", blog.RegisterUserHandler(svc))
	router.Handler(http.MethodPost, "/v1/auth/login", blog.LoginHandler(svc))
	router.Handler(http.MethodPost, "/v1/posts", blog.RequireAuth(blog.CreatePostHandler(svc)))
	router.Handler(http.MethodGet, "/v1/users/:username", blog.GetProfileHandler(svc))

	log.Printf("Server started. Listening on port: %s\n", "8090")
	log.Fatal(http.ListenAndServe(":"+("8090"), router))
}
