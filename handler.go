package gomicroblog

import (
	"encoding/json"
	"net/http"
)

type registerUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func decodeRequest(ur registerUserRequest, r *http.Request) (registerUserRequest, error) {
	if err := json.NewDecoder(r.Body).Decode(&ur); err != nil {
		return registerUserRequest{}, err
	}

	return ur, nil
}
