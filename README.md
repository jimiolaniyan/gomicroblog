# Go Microblog

This project demonstrates building [Miguel Grinberg's](https://github.com/miguelgrinberg/microblog) Microblog tutorial
application in Golang using TDD, BDD and Robert C. Martin's Clean Architecture.

## Getting started
- Clone the repo  
`git clone https://github.com/jimiolaniyan/gomicroblog.git`  
`cd gomicroblog`
 
- Run the tests  
`go test ./... -v -short`  
 Note: running without `-short` flag requires `docker` with `mongo:latest` image
- Build the project  
`go build -o blog api/main.go`
- Start the server  
`./blog`

### Run in docker container
Within the project directory:  
`docker image build -t microblog:0.3.0 .`  
`docker container run -p 8090:8090 --name microblog microblog:0.3.0`

## API Requests 
See `api/requests.http` for full examples.
### Register new user
```
curl -X POST \
  http://localhost:8090/v1/users \
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
  http://localhost:8090/v1/sessions \
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

