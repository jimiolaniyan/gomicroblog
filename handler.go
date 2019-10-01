package gomicroblog

import (
	"encoding/json"
	"fmt"
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
			encodeError(err, w)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.String(), id))
		json.NewEncoder(w).Encode(&registerUserResponse{ID: id})
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func decodeRegisterUserRequest(r *http.Request) (registerUserRequest, error) {
	ur := registerUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&ur); err != nil {
		return registerUserRequest{}, err
	}

	return ur, nil
}
