package gomicroblog

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeRequest(t *testing.T) {
	registerReq := `{ "username": "jimi", "password": "password1", "email": "test@tester.test" }`
	loginReq := `{"username": "jimi", "password": "password1"}`
	createPostReq := `{"body": "a simple post"}`
	tests := []struct {
		r       string
		decoder func(closer io.ReadCloser) (interface{}, error)
		wantErr error
		wantReq interface{}
	}{
		{registerReq, decodeRegisterUserRequest, nil, registerUserRequest{"jimi", "password1", "test@tester.test"}},
		{loginReq, decodeValidateUserRequest, nil, validateUserRequest{"jimi", "password1"}},
		{createPostReq, decodeCreatePostRequest, nil, createPostRequest{"a simple post"}},
	}

	for _, tt := range tests {
		body := ioutil.NopCloser(strings.NewReader(tt.r))
		req, err := tt.decoder(body)
		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantReq, req)
	}
}

func TestHandlerResponses(t *testing.T) {
	svc := &service{users: NewUserRepository()}
	url := "/v1/users/new"
	registerHandler := RegisterUserHandler(svc)
	registerReq := `
		{
			"username":"jimi",
			"password":"password1",
			"email":"test@tester.test"
		}
`
	tests := []struct {
		method, req  string
		wantCode     int
		wantValidID  bool
		wantErr      error
		wantLocation string
		testExisting bool
	}{
		{
			http.MethodPost,
			registerReq,
			http.StatusCreated,
			true,
			errors.New(""),
			"/v1/users",
			false,
		},
		{
			http.MethodPost,
			`invalid request`,
			http.StatusBadRequest,
			false,
			errors.New(""),
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "", "password": "pass"}`,
			http.StatusUnprocessableEntity,
			false,
			ErrInvalidUsername,
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "username", "password": "pass", "email": "a@b.com"}`,
			http.StatusUnprocessableEntity,
			false,
			ErrInvalidPassword,
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "username", "password": "password", "email": "ab.com"}`,
			http.StatusUnprocessableEntity,
			false,
			ErrInvalidEmail,
			"",
			false,
		},
		{
			http.MethodPost,
			`{"username": "jimi", "password": "password", "email": "a@b.com"}`,
			http.StatusConflict,
			false,
			ErrExistingUsername,
			"",
			true,
		},
		{
			http.MethodPost,
			`{"username": "username", "password": "password", "email": "test@tester.test"}`,
			http.StatusConflict,
			false,
			ErrExistingEmail,
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.wantCode), func(t *testing.T) {
			if tt.testExisting {
				ur := registerUserRequest{}
				req := registerReq
				_ = json.NewDecoder(strings.NewReader(req)).Decode(&ur)
				user, _ := NewUser(ur.Username, ur.Email)
				_ = svc.users.Store(user)
			}

			r, err := http.NewRequest(tt.method, url, strings.NewReader(tt.req))
			assert.Nil(t, err)

			w := httptest.NewRecorder()
			handler := http.NewServeMux()
			handler.Handle(url, registerHandler)
			handler.ServeHTTP(w, r)

			var res struct {
				ID  ID     `json:"id,omitempty"`
				Err string `json:"error,omitempty"`
			}

			_ = json.NewDecoder(w.Body).Decode(&res)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantErr.Error(), res.Err)
			assert.Equal(t, IsValidID(string(res.ID)), tt.wantValidID)
			assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
			assert.True(t, strings.HasPrefix(w.Header().Get("Location"), tt.wantLocation))
		})
	}
}

func TestLoginHandler(t *testing.T) {
	svc := NewService(NewUserRepository(), nil)
	userID, _ := svc.RegisterNewUser(registerUserRequest{"user", "password", "a@b.com"})

	tests := []struct {
		description, req       string
		wantCode, wantTokenLen int
		wantClaims             string
	}{
		{"BadRequest", `invalid request`, http.StatusBadRequest, 1, ""},
		{"NonExistentUser", `{"username": "nonexistent", "password": "password"}`, http.StatusUnauthorized, 1, ""},
		{"ExistingUserWithInvalidPassword", `{"username": "user", "password": "anInvalid"}`, http.StatusUnauthorized, 1, ""},
		{"ExistingUserWithValidPassword", `{"username": "user", "password": "password"}`, http.StatusOK, 3, fmt.Sprintf("{\"iss\":\"auth\",\"sub\":\"%s\"}", userID)},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(tt.req))
			assert.Nil(t, err)

			w := httptest.NewRecorder()
			mux := http.NewServeMux()
			mux.Handle("/v1/auth/login", LoginHandler(svc))
			mux.ServeHTTP(w, req)

			var res struct {
				Token string `json:"token,omitempty"`
				Error string `json:"error,omitempty"`
			}

			_ = json.NewDecoder(w.Body).Decode(&res)
			assert.Equal(t, tt.wantCode, w.Code)
			parts := strings.Split(res.Token, ".")
			assert.Equal(t, len(parts), tt.wantTokenLen)

			if len(parts) > 2 {
				claim, err := base64.RawStdEncoding.DecodeString(parts[1])
				assert.Nil(t, err)
				assert.Equal(t, tt.wantClaims, string(claim))
			}
		})
	}
}

func TestCreatePostHandler(t *testing.T) {
	svc := NewService(NewUserRepository(), NewPostRepository())
	id, _ := svc.RegisterNewUser(registerUserRequest{"user", "password", "a@b.com"})

	tests := []struct {
		req, userID  string
		wantCode     int
		wantErr      error
		wantID       bool
		wantLocation string
		wantCtx      bool
	}{
		{`invalid request`, "", http.StatusBadRequest, errors.New(""), false, "", true},
		{`{}`, "", http.StatusInternalServerError, ErrEmptyContext, false, "", false},
		{`{"body": ""}`, "", http.StatusUnauthorized, ErrInvalidID, false, "", true},
		{`{"body": ""}`, "puoiwoerigp", http.StatusUnauthorized, ErrInvalidID, false, "", true},
		{`{"body": ""}`, string(id), http.StatusUnprocessableEntity, ErrEmptyBody, false, "", true},
		{`{"body": "i love my wife :)"}`, string(id), http.StatusCreated, errors.New(""), true, "/v1/posts/", true},
	}

	for _, tt := range tests {
		r, _ := http.NewRequest(http.MethodPost, "/v1/posts", strings.NewReader(tt.req))

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.Handle("/v1/posts", CreatePostHandler(svc))

		if tt.wantCtx {
			ctx := context.WithValue(r.Context(), idKey, tt.userID)
			r = r.WithContext(ctx)
		}

		mux.ServeHTTP(w, r)

		var res struct {
			ID  PostID `json:"id,omitempty"`
			Err string `json:"error,omitempty"`
		}

		_ = json.NewDecoder(w.Body).Decode(&res)

		assert.Equal(t, tt.wantCode, w.Code)
		assert.Equal(t, tt.wantErr.Error(), res.Err)
		assert.Equal(t, IsValidID(string(res.ID)), tt.wantID)
		assert.True(t, strings.HasPrefix(w.Header().Get("Location"), tt.wantLocation))
	}

}

func TestRequireAuth(t *testing.T) {
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU"
	validToken, _ := getJWTToken("randomid")

	tests := []struct {
		authHeader string
		wantCode   int
		wantID     string
	}{
		{authHeader: "", wantCode: http.StatusUnauthorized},
		{authHeader: "k", wantCode: http.StatusUnauthorized},
		{authHeader: "Random random", wantCode: http.StatusUnauthorized},
		{authHeader: "Bearer ", wantCode: http.StatusUnauthorized},
		{authHeader: "Bearer random.random.random", wantCode: http.StatusUnauthorized},
		{authHeader: "Bearer " + invalidToken, wantCode: http.StatusUnauthorized},
		{authHeader: "Bearer " + validToken, wantCode: http.StatusOK, wantID: "randomid"},
	}

	for _, tt := range tests {
		var id string
		f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id = r.Context().Value(idKey).(string)
			return
		})

		h := RequireAuth(f)
		r, _ := http.NewRequest(http.MethodPost, "/v1/posts", nil)
		r.Header.Set("Authorization", tt.authHeader)

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.Handle("/v1/posts", h)

		mux.ServeHTTP(w, r)

		assert.IsType(t, new(http.Handler), &h)
		assert.Equal(t, tt.wantID, id)
		assert.Equal(t, tt.wantCode, w.Code)
	}
}
