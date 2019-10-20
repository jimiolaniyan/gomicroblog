package gomicroblog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/dgrijalva/jwt-go"
)

type key string

const idKey = key("auth")

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
			return
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

var ErrEmptyContext = errors.New("could not get user id from context")

func CreatePostHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeCreatePostRequest(r.Body)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			encodeError(ErrEmptyContext, w)
			return
		}

		req := request.(createPostRequest)
		postId, err := svc.CreatePost(ID(userID), req.Body)

		if err != nil {
			encodeError(err, w)
			return
		}

		location := strings.Join(strings.Split(r.URL.Path, "/")[0:3], "/")
		w.Header().Set("Location", fmt.Sprintf("%s/%s", location, postId))
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(&createPostResponse{ID: postId}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func GetProfileHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := getNameFromRequestParams(r, "username")

		w.Header().Set("Content-Type", "application/json")
		if username == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		p, err := svc.GetProfile(username)
		if err != nil {
			encodeError(err, w)
			return
		}

		if err = json.NewEncoder(w).Encode(profileResponse{Profile: &p, URL: r.URL.String()}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

	})
}

func EditProfileHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeEditProfileRequest(r.Body)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id, ok := getUserIDFromContext(r.Context())
		if !ok {
			encodeError(ErrEmptyContext, w)
			return
		}

		err = svc.EditProfile(ID(id), request.(editProfileRequest))
		if err != nil {
			encodeError(err, w)
			return
		}
	})
}

func CreateRelationshipHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		username, id, ok := getRelationshipRequestParams(r, w)
		if !ok {
			return
		}

		if err := svc.CreateRelationshipFor(ID(id), username); err != nil {
			encodeError(err, w)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	})
}

func RemoveRelationshipHandler(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, id, ok := getRelationshipRequestParams(r, w)
		if !ok {
			return
		}

		if err := svc.RemoveRelationshipFor(ID(id), username); err != nil {
			encodeError(err, w)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	})
}

func getRelationshipRequestParams(r *http.Request, w http.ResponseWriter) (string, string, bool) {
	username := getNameFromRequestParams(r, "username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return "", "", false
	}

	id, ok := getUserIDFromContext(r.Context())
	if !ok {
		encodeError(ErrEmptyContext, w)
		return "", "", false
	}

	return username, id, true
}

func LastSeenMiddleware(f http.Handler, svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := getUserIDFromContext(r.Context())
		if !ok {
			encodeError(ErrEmptyContext, w)
			return
		}

		err := svc.UpdateLastSeen(ID(id))
		if err != nil {
			encodeError(err, w)
			return
		}
		f.ServeHTTP(w, r)
	})
}

func RequireAuth(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := getTokenFromRequest(r)

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("error verifying signing method")
			}
			return []byte("e624d92e3fa438b6a8fac4f698e977cd"), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		ctx := context.WithValue(r.Context(), idKey, claims["sub"])

		f.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getNameFromRequestParams(r *http.Request, name string) string {
	params := httprouter.ParamsFromContext(r.Context())
	username := strings.TrimSpace(params.ByName(name))
	return username
}

func getUserIDFromContext(ctx context.Context) (id string, ok bool) {
	id, ok = ctx.Value(idKey).(string)
	return
}

func getTokenFromRequest(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToUpper(parts[0]) != "BEARER" {
		return ""
	}
	return parts[1]
}

func getJWTToken(id string) (string, error) {
	key := []byte("e624d92e3fa438b6a8fac4f698e977cd")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{Issuer: "auth", Subject: id})
	return token.SignedString(key)
}

func encodeError(err error, w http.ResponseWriter) {
	switch err {
	case ErrInvalidCredentials, ErrInvalidID:
		w.WriteHeader(http.StatusUnauthorized)
	case ErrCantFollowSelf, ErrCantUnFollowSelf:
		w.WriteHeader(http.StatusForbidden)
	case ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case ErrExistingUsername, ErrExistingEmail, ErrAlreadyFollowing, ErrNotFollowing:
		w.WriteHeader(http.StatusConflict)
	case ErrEmptyBody, ErrInvalidEmail, ErrInvalidPassword, ErrInvalidUsername, ErrBioTooLong:
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

func decodeCreatePostRequest(body io.ReadCloser) (interface{}, error) {
	req := createPostRequest{}
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return createPostRequest{}, err
	}
	return req, nil
}

func decodeEditProfileRequest(body io.ReadCloser) (interface{}, error) {
	req := editProfileRequest{}
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return editProfileRequest{}, err
	}
	return req, nil
}
