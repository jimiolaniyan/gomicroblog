package gomicroblog

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io"
	"net/http"
	"strings"
)

func RegisterUserHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRegisterUserRequest(r.Body)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := svc.RegisterNewUser(req.(registerUserRequest))
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

func LoginHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeValidateUserRequest(r.Body)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		userID, err := svc.ValidateUser(req.(validateUserRequest))
		if err != nil {
			encodeError(err, w)
			return
		}

		tokenString, err := getJWTToken(string(userID))
		if err = json.NewEncoder(w).Encode(map[string]interface{}{"token": tokenString}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func getJWTToken(id string) (string, error) {
	key := []byte("e624d92e3fa438b6a8fac4f698e977cd")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{Issuer: "auth", Subject: id})
	return token.SignedString(key)
}

func encodeError(err error, w http.ResponseWriter) {
	switch err {
	case ErrExistingUsername, ErrExistingEmail:
		w.WriteHeader(http.StatusConflict)
	case ErrInvalidEmail, ErrInvalidPassword, ErrInvalidUsername:
		w.WriteHeader(http.StatusUnprocessableEntity)
	case ErrInvalidCredentials:
		w.WriteHeader(http.StatusUnauthorized)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func decodeRegisterUserRequest(body io.ReadCloser) (interface{}, error) {
	req := registerUserRequest{}
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return registerUserRequest{}, err
	}

	return req, nil
}

func decodeValidateUserRequest(body io.ReadCloser) (interface{}, error) {
	req := validateUserRequest{}
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return validateUserRequest{}, err
	}
	return req, nil
}
