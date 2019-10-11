package main

import (
	blog "github.com/jimiolaniyan/gomicroblog"
	"log"
	"net/http"
)

func main() {
	users := blog.NewUserRepository()
	posts := blog.NewPostRepository()
	svc := blog.NewService(users, posts)

	mux := http.NewServeMux()
	mux.Handle("/v1/users/new", blog.RegisterUserHandler(svc))
	mux.Handle("/v1/auth/login", blog.LoginHandler(svc))
	mux.Handle("/v1/posts", blog.RequireAuth(blog.CreatePostHandler(svc)))

	port := "8090"
	log.Printf("Listening on port: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
