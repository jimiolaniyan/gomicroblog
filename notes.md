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
* run `go mod init github.com/<jimiolaniyan>/gomicrolog`
* Add import `. "github.com/smartystreets/goconvey/convey"`
* run `go test ./...` to get convey dependency


Handler-based Architecture:
* Handler decodes transport request and creates a request model
* Handler then invokes the use case with a request model and responder `service.RegisterNewUser(req, res)`
* Use case generates the output and hands it to the responder `responder.Format(om)`
* Formatter generates response model ``
* Handler responds to the transport with the  

Go kit arch:
* Handler decodes request and calls the use case with it
* Use case generates the output and returns it to the handler
* The handler creates a response model, encodes it and returns it to the client 
 