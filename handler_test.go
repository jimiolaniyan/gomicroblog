package gomicroblog

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/assert"
)

type HandlerTestSuite struct {
	suite.Suite
	userID ID
	svc    Service
	req    registerUserRequest
}

func (hs *HandlerTestSuite) SetupSuite() {
	hs.svc = NewService(NewUserRepository(), NewPostRepository())
	hs.req = registerUserRequest{"user", "password", "a@b.com"}
	id, _ := hs.svc.RegisterNewUser(hs.req)
	hs.userID = id
}

func TestDecodeRequest(t *testing.T) {
	registerReq := `{ "username": "jimi", "password": "password1", "email": "test@tester.test" }`
	loginReq := `{"username": "jimi", "password": "password1"}`
	createPostReq := `{"body": "a simple post"}`
	u := "new"
	tests := []struct {
		r       string
		decoder func(closer io.ReadCloser) (interface{}, error)
		wantErr error
		wantReq interface{}
	}{
		{registerReq, decodeRegisterUserRequest, nil, registerUserRequest{"jimi", "password1", "test@tester.test"}},
		{loginReq, decodeValidateUserRequest, nil, validateUserRequest{"jimi", "password1"}},
		{createPostReq, decodeCreatePostRequest, nil, createPostRequest{"a simple post"}},
		{`{}`, decodeEditProfileRequest, nil, editProfileRequest{nil, nil}},
		{`{"username": "", "bio": ""}`, decodeEditProfileRequest, nil, editProfileRequest{new(string), new(string)}},
		{`{"username": "new", "bio": "new"}`, decodeEditProfileRequest, nil, editProfileRequest{&u, &u}},
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
		{http.MethodPost,
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

func (hs *HandlerTestSuite) TestLoginHandler() {
	validClaims := fmt.Sprintf("{\"iss\":\"auth\",\"sub\":\"%s\"}", hs.userID)
	tests := []struct {
		description, req       string
		wantCode, wantTokenLen int
		wantClaims             string
	}{
		{"BadRequest", `invalid request`, http.StatusBadRequest, 1, ""},
		{"NonExistentUser", `{"username": "nonexistent", "password": "password"}`, http.StatusUnauthorized, 1, ""},
		{"ExistingUserWithInvalidPassword", `{"username": "user", "password": "anInvalid"}`, http.StatusUnauthorized, 1, ""},
		{"ExistingUserWithValidPassword", `{"username": "user", "password": "password"}`, http.StatusOK, 3, validClaims},
	}

	for _, tt := range tests {
		hs.Run(tt.description, func() {
			req, err := http.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(tt.req))
			assert.Nil(hs.T(), err)

			w := httptest.NewRecorder()
			mux := http.NewServeMux()
			mux.Handle("/v1/auth/login", LoginHandler(hs.svc))
			mux.ServeHTTP(w, req)

			var res struct {
				Token string `json:"token,omitempty"`
				Error string `json:"error,omitempty"`
			}

			_ = json.NewDecoder(w.Body).Decode(&res)
			assert.Equal(hs.T(), tt.wantCode, w.Code)
			parts := strings.Split(res.Token, ".")
			assert.Equal(hs.T(), len(parts), tt.wantTokenLen)

			if len(parts) > 2 {
				claim, err := base64.RawStdEncoding.DecodeString(parts[1])
				assert.Nil(hs.T(), err)
				assert.Equal(hs.T(), tt.wantClaims, string(claim))
			}
		})
	}
}

func (hs *HandlerTestSuite) TestCreatePostHandler() {
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
		{`{"body": ""}`, string(hs.userID), http.StatusUnprocessableEntity, ErrEmptyBody, false, "", true},
		{`{"body": "i love my wife :)"}`, string(hs.userID), http.StatusCreated, errors.New(""), true, "/v1/posts/", true},
	}

	for _, tt := range tests {
		r, _ := http.NewRequest(http.MethodPost, "/v1/posts", strings.NewReader(tt.req))

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.Handle("/v1/posts", CreatePostHandler(hs.svc))

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

		assert.Equal(hs.T(), tt.wantCode, w.Code)
		assert.Equal(hs.T(), tt.wantErr.Error(), res.Err)
		assert.Equal(hs.T(), IsValidID(string(res.ID)), tt.wantID)
		assert.True(hs.T(), strings.HasPrefix(w.Header().Get("Location"), tt.wantLocation))
	}

}

func (hs *HandlerTestSuite) TestGetProfileHandler() {
	u := hs.req.Username
	host := "http://localhost:8080"
	finalURL := fmt.Sprintf("%s/v1/users/%s", host, u)
	nilErr := errors.New("")

	tests := []struct {
		username              string
		wantCode              int
		wantErr               error
		wantID                bool
		wantUsername, wantURL string
	}{
		{username: "  ", wantCode: http.StatusBadRequest, wantErr: nilErr, wantUsername: ""},
		{username: "nonexistent", wantCode: http.StatusNotFound, wantErr: ErrNotFound, wantUsername: ""},
		{username: u, wantCode: http.StatusOK, wantErr: nilErr, wantID: true, wantUsername: u, wantURL: finalURL},
	}

	for _, tt := range tests {
		url := fmt.Sprintf("%s/v1/users/%s", host, tt.username)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		router := httprouter.New()

		router.Handler(http.MethodGet, "/v1/users/:username", GetProfileHandler(hs.svc))
		router.ServeHTTP(w, req)

		var res struct {
			Profile Profile `json:"profile,omitempty"`
			URL     string  `json:"url,omitempty"`
			Err     string  `json:"error,omitempty"`
		}

		_ = json.NewDecoder(w.Body).Decode(&res)

		assert.Equal(hs.T(), tt.wantCode, w.Code)
		assert.Equal(hs.T(), tt.wantErr.Error(), res.Err)
		assert.Equal(hs.T(), tt.wantID, IsValidID(string(res.Profile.ID)))
		assert.Equal(hs.T(), tt.wantUsername, res.Profile.Username)
		assert.Equal(hs.T(), tt.wantURL, res.URL)
	}

}

func (hs *HandlerTestSuite) TestEditProfileHandler() {
	// register new user to avoid conflicts with user in the suite
	req := registerUserRequest{"tempUser", "password", "temp@mail.com"}
	id, _ := hs.svc.RegisterNewUser(req)
	sid := string(id)

	nilErr := errors.New("")
	longBio := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut " +
		"labore et dolore magna aliqua. Ut enim ad minim h"

	S422 := http.StatusUnprocessableEntity

	tests := []struct {
		req                   string
		id                    string
		withCtx, reset        bool
		wantCode              int
		wantErr               error
		wantUsername, wantBio string
	}{
		{req: `invalid request`, wantCode: http.StatusBadRequest, wantErr: nilErr},
		{req: `{}`, wantCode: http.StatusInternalServerError, wantErr: ErrEmptyContext},
		{req: `{}`, id: "invalid", wantCode: http.StatusUnauthorized, withCtx: true, wantErr: ErrInvalidID},
		{req: `{}`, id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: nilErr},
		{req: `{"username": ""}`, id: sid, wantCode: S422, withCtx: true, wantErr: ErrInvalidUsername},
		{req: fmt.Sprintf(`{"username": "%s"}`, req.Username), id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: nilErr},
		{req: fmt.Sprintf(`{"username": "%s"}`, hs.req.Username), id: sid, wantCode: http.StatusConflict, withCtx: true, wantErr: ErrExistingUsername},
		{req: `{"username": "newName"}`, id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: nilErr, wantUsername: "newName", reset: true},
		{req: `{"bio": ""}`, id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: nilErr},
		{req: fmt.Sprintf(`{"bio": "%s"}`, longBio), id: sid, wantCode: S422, withCtx: true, wantErr: ErrBioTooLong},
		{req: `{"bio": "Bios"}`, id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: nilErr, wantBio: "Bios"},
		{req: `{"username": "newU", "bio": "Be nice"}`, id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: nilErr, wantBio: "Be nice", wantUsername: "newU"},
	}

	for _, tt := range tests {
		r, _ := http.NewRequest(http.MethodPatch, "/v1/users", strings.NewReader(tt.req))

		u := req.Username
		if len(tt.wantUsername) > 0 {
			u = tt.wantUsername
		}

		r2, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/users/%s", u), nil)

		if tt.withCtx {
			r = r.WithContext(context.WithValue(r.Context(), idKey, tt.id))
		}

		router := httprouter.New()
		router.Handler(http.MethodPatch, "/v1/users", EditProfileHandler(hs.svc))
		router.Handler(http.MethodGet, "/v1/users/:username", GetProfileHandler(hs.svc))

		w := httptest.NewRecorder()
		w2 := httptest.NewRecorder()

		router.ServeHTTP(w, r)
		router.ServeHTTP(w2, r2)

		var res struct {
			Err string `json:"error"`
		}

		var res2 struct {
			Profile Profile `json:"profile,omitempty"`
		}

		_ = json.NewDecoder(w.Body).Decode(&res)
		_ = json.NewDecoder(w2.Body).Decode(&res2)

		assert.Equal(hs.T(), tt.wantCode, w.Code)
		assert.Equal(hs.T(), tt.wantErr.Error(), res.Err)
		assert.Equal(hs.T(), tt.wantBio, res2.Profile.Bio)

		if len(tt.wantUsername) > 0 {
			assert.Equal(hs.T(), tt.wantUsername, res2.Profile.Username)
		} else {
			assert.Equal(hs.T(), req.Username, res2.Profile.Username)
		}

		//reset the username and bio
		if tt.reset {
			body := strings.NewReader(fmt.Sprintf(`{"username": "%s", "bio": ""}`, req.Username))
			r3, _ := http.NewRequest(http.MethodPatch, "/v1/users", body)
			r3 = r3.WithContext(context.WithValue(r3.Context(), idKey, tt.id))
			router.ServeHTTP(w, r3)
		}
	}
}

func TestRequireAuthMiddleware(t *testing.T) {
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

	var id string
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id = r.Context().Value(idKey).(string)
		return
	})

	for _, tt := range tests {

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

func (hs *HandlerTestSuite) TestLastSeenMiddleware() {
	now := time.Now().UTC()
	var called bool
	f := func(w http.ResponseWriter, r *http.Request) {
		if id, _ := getUserIDFromContext(r.Context()); id != "" {
			called = true
		}
	}

	l := LastSeenMiddleware(http.HandlerFunc(f), hs.svc)
	r, _ := http.NewRequest("", "/doesnt-matter", nil)
	pr, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/users/%s", hs.req.Username), nil)

	tests := []struct {
		id                          string
		wantCode                    int
		withCtx, wantLS, wantCalled bool
	}{
		{wantCode: http.StatusInternalServerError},
		{id: "invalid", withCtx: true, wantCode: http.StatusNotFound},
		{id: string(hs.userID), withCtx: true, wantCode: http.StatusOK, wantLS: true, wantCalled: true},
	}

	for _, tt := range tests {
		if tt.withCtx {
			r = r.WithContext(context.WithValue(r.Context(), idKey, tt.id))
		}

		w := httptest.NewRecorder()
		router := httprouter.New()
		router.Handler(http.MethodGet, "/doesnt-matter", l)
		router.Handler(http.MethodGet, "/v1/users/:username", GetProfileHandler(hs.svc))
		router.ServeHTTP(w, r)
		router.ServeHTTP(w, pr)

		var res struct {
			Profile Profile `json:"profile,omitempty"`
		}

		_ = json.NewDecoder(w.Body).Decode(&res)

		assert.IsType(hs.T(), new(http.Handler), &l)
		assert.Equal(hs.T(), tt.wantCode, w.Code)
		assert.Equal(hs.T(), tt.wantCalled, called)
		assert.Equal(hs.T(), tt.wantLS, res.Profile.LastSeen.After(now))
	}
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
