package gomicroblog

import (
	"encoding/json"
	"net/http"
)

func RegisterUserHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rm, err := decodeRegisterUserRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := svc.RegisterNewUser(rm)
		if err != nil {
			res := struct {
				Err string `json:"err"`
			}{
				Err: err.Error(),
			}
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(&res)
		}
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
