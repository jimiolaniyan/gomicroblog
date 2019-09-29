package gomicroblog

import (
	"encoding/json"
	"net/http"
)

func RegisterUserHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rm, _ := decodeRegisterUserRequest(r)
		id, _ := svc.RegisterNewUser(rm)
		json.NewEncoder(w).Encode(&registerUserResponse{ID: id})
	})
}

func decodeRegisterUserRequest(r *http.Request) (registerUserRequest, error) {
	ur := registerUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&ur); err != nil {
		return registerUserRequest{}, err
	}

	return ur, nil
}
