package main

import (
	blog "github.com/jimiolaniyan/gomicroblog"
	"log"
	"net/http"
)

func main() {
	users := blog.NewUserRepository()
	svc := blog.NewService(users)

	mux := http.NewServeMux()
	mux.Handle("/users/v1/new", blog.RegisterUserHandler(svc))
	port := "8090"
	log.Printf("Listening on port: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
