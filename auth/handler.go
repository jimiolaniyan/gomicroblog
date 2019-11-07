package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var signingKey = []byte(os.Getenv("AUTH_SIGNING_KEY"))

func RegisterAccountHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRegisterAccountRequest(r.Body)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id, err := svc.RegisterAccount(req)
		if err != nil {
			encodeError(err, w)
			return
		}

		loc := strings.Join(strings.Split(r.URL.Path, "/")[0:4], "/")

		w.Header().Set("Location", fmt.Sprintf("%s/%s", loc, id))
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(registerAccountResponse{ID: id}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func LoginHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeLoginRequest(r.Body)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id, err := svc.ValidateCredentials(req)
		if err != nil {
			encodeError(err, w)
			return
		}

		tokenString, err := getJWTToken(string(id))
		if err = json.NewEncoder(w).Encode(map[string]interface{}{"token": tokenString}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func getJWTToken(id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{Issuer: "auth", Subject: id})
	return token.SignedString(signingKey)
}

func encodeError(err error, w http.ResponseWriter) {
	switch err {
	case ErrInvalidCredentials:
		w.WriteHeader(http.StatusUnauthorized)
	case ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case ErrExistingUsername, ErrExistingEmail:
		w.WriteHeader(http.StatusConflict)
	case ErrInvalidEmail, ErrInvalidPassword, ErrInvalidUsername:
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

func decodeRegisterAccountRequest(body io.ReadCloser) (registerAccountRequest, error) {
	req := registerAccountRequest{}
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return registerAccountRequest{}, err
	}
	return req, nil
}

func decodeLoginRequest(body io.ReadCloser) (validateCredentialsRequest, error) {
	req := validateCredentialsRequest{}
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return validateCredentialsRequest{}, err
	}
	return req, nil
}
