package gomicroblog

import (
	"encoding/json"
	"net/http"
)

type Responder interface {
}

func RegisterUserHandler(svc Service, res Responder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rm, _ := decodeRegisterUserRequest(r)
		svc.RegisterNewUser(rm, res)
	})
}

func decodeRegisterUserRequest(r *http.Request) (registerUserRequest, error) {
	ur := registerUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&ur); err != nil {
		return registerUserRequest{}, err
	}

	return ur, nil
}
