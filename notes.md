#Go Microblog

## Steps:
* Install convey
* Run convey $GOPATH/bin/convey
* Compose:
```
Given new user with username, email and password
	When user creates account
		Then user account is created
```
* Create `bdd_test`
* copy composer output to `bdd_test` 
* run `go mod init github.com/jimiolaniyan/gomicrolog`
* Add import `. "github.com/smartystreets/goconvey/convey"`
* run `go test ./...` to get convey dependency