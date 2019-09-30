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

## API Requests
### Register new user
```
curl -X GET \
  http://localhost:8090/users/v1/new \
  -H 'Content-Type: application/json' \
  -d '{
	"username": "jimi",
	"password": "mypassword",
	"email": "a@b.com"
}'
```


