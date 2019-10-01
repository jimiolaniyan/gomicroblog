package gomicroblog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func RegisterUserHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rm, err := decodeRegisterUserRequest(r)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := svc.RegisterNewUser(rm)
		if err != nil {
			encodeError(err, w)
			return
		}
		location := strings.Join(strings.Split(r.URL.Path, "/")[0:3], "/")
		w.Header().Set("Location", fmt.Sprintf("%s/%s", location, id))
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(&registerUserResponse{ID: id}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func encodeError(err error, w http.ResponseWriter) {
	switch err {
	case ErrExistingUsername, ErrExistingEmail:
		w.WriteHeader(http.StatusConflict)
	case ErrInvalidEmail, ErrInvalidPassword, ErrEmptyUserName:
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func decodeRegisterUserRequest(r *http.Request) (registerUserRequest, error) {
	ur := registerUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&ur); err != nil {
		return registerUserRequest{}, err
	}

	return ur, nil
}
