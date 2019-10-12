# Go Microblog

This project demonstrates building [Miguel Grinberg's](https://github.com/miguelgrinberg/microblog) Microblog tutorial
application in Golang using TDD, BDD and Robert C. Martin's Clean Architecture.

## Getting started
- Clone the repo  
`git clone https://github.com/jimiolaniyan/gomicroblog.git`  
`cd gomicroblog`
 
- Run the tests  
`go test -v ./...`
- Build the project  
`go build -o blog api/main.go`
- Run the app with  
`./blog`

## API Requests 
See `api/requests.http` for full examples.
### Register new user
```
curl -X POST \
  http://localhost:8090/v1/users/new \
  -H 'Content-Type: application/json' \
  -d '{
	"username": "jimi",
	"password": "mypassword",
	"email": "a@b.com"
}'
```
### Login existing user
```
curl -X POST \
  http://localhost:8090/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{ 
        "username": "jimi",
        "password": "mypassword"
}'
```
### Create a new post
```
curl -X POST \
  http://localhost:8090/v1/posts \
  -H 'Authorization: Bearer <token.goes.here>'
  -H 'Content-Type: application/json'
  -d '{"body": "a post"}'
```

