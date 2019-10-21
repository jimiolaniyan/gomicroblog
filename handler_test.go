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

var nilErr = errors.New("")

func TestRegisterNewUserHandler(t *testing.T) {
	svc := &service{users: NewUserRepository()}
	url := "/v1/users"
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
			nilErr,
			"/v1/users",
			false,
		},
		{
			http.MethodPost,
			`invalid request`,
			http.StatusBadRequest,
			false,
			nilErr,
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
			req, err := http.NewRequest(http.MethodPost, "/v1/sessions", strings.NewReader(tt.req))
			assert.Nil(hs.T(), err)

			w := httptest.NewRecorder()
			mux := http.NewServeMux()
			mux.Handle("/v1/sessions", LoginHandler(hs.svc))
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
		{`invalid request`, "", http.StatusBadRequest, nilErr, false, "", true},
		{`{}`, "", http.StatusInternalServerError, ErrEmptyContext, false, "", false},
		{`{"body": ""}`, "", http.StatusUnauthorized, ErrInvalidID, false, "", true},
		{`{"body": ""}`, "puoiwoerigp", http.StatusUnauthorized, ErrInvalidID, false, "", true},
		{`{"body": ""}`, string(hs.userID), http.StatusUnprocessableEntity, ErrEmptyBody, false, "", true},
		{`{"body": "i love my wife :)"}`, string(hs.userID), http.StatusCreated, nilErr, true, "/v1/posts/", true},
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
			r = setIDInRequestContext(r, tt.id)
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

func (hs *HandlerTestSuite) TestCreateRelationshipHandler() {
	uid := string(hs.userID)

	u := "followUser"
	_, err := hs.svc.RegisterNewUser(registerUserRequest{u, "password", "f@u.co"})
	assert.Nil(hs.T(), err)

	tests := []struct {
		username string
		withCtx  bool
		wantCode int
		wantErr  error
	}{
		{username: "  ", wantCode: http.StatusBadRequest, wantErr: nilErr},
		{username: "nonexistent", wantCode: http.StatusInternalServerError, wantErr: ErrEmptyContext},
		{username: "nonexistent", withCtx: true, wantCode: http.StatusNotFound, wantErr: ErrNotFound},
		{username: hs.req.Username, withCtx: true, wantCode: http.StatusForbidden, wantErr: ErrCantFollowSelf},
		{username: u, withCtx: true, wantCode: http.StatusNoContent, wantErr: nilErr},
		{username: u, withCtx: true, wantCode: http.StatusConflict, wantErr: ErrAlreadyFollowing},
	}

	for _, tt := range tests {
		u := fmt.Sprintf("/v1/users/%s/followers", tt.username)
		r, _ := http.NewRequest(http.MethodPost, u, nil)

		if tt.withCtx {
			r = setIDInRequestContext(r, uid)
		}

		router := httprouter.New()
		router.Handler(http.MethodPost, "/v1/users/:username/followers", CreateRelationshipHandler(hs.svc))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		var res struct {
			Err string `json:"error,omitempty"`
		}

		_ = json.NewDecoder(w.Body).Decode(&res)

		assert.Equal(hs.T(), tt.wantCode, w.Code)
		assert.Equal(hs.T(), tt.wantErr.Error(), res.Err)
	}

	// clean up
	_ = hs.svc.RemoveRelationshipFor(ID(uid), u)
}

func (hs *HandlerTestSuite) TestRemoveRelationshipHandler() {
	uid := string(hs.userID)

	u := "unFollowUser"
	_, _ = hs.svc.RegisterNewUser(registerUserRequest{u, "password", "u@u.co"})
	_ = hs.svc.CreateRelationshipFor(ID(uid), u)

	tests := []struct {
		username string
		withCtx  bool
		wantCode int
		wantErr  error
	}{
		{username: "  ", wantCode: http.StatusBadRequest, wantErr: nilErr},
		{username: "nonexistent", wantCode: http.StatusInternalServerError, wantErr: ErrEmptyContext},
		{username: "nonexistent", withCtx: true, wantCode: http.StatusNotFound, wantErr: ErrNotFound},
		{username: hs.req.Username, withCtx: true, wantCode: http.StatusForbidden, wantErr: ErrCantUnFollowSelf},
		{username: u, withCtx: true, wantCode: http.StatusNoContent, wantErr: nilErr},
		{username: u, withCtx: true, wantCode: http.StatusConflict, wantErr: ErrNotFollowing},
	}

	for _, tt := range tests {
		r, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/users/%s/followers", tt.username), nil)

		if tt.withCtx {
			r = setIDInRequestContext(r, uid)
		}

		router := httprouter.New()
		router.Handler(http.MethodDelete, "/v1/users/:username/followers", RemoveRelationshipHandler(hs.svc))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		var res struct {
			Err string `json:"error,omitempty"`
		}

		_ = json.NewDecoder(w.Body).Decode(&res)

		assert.Equal(hs.T(), tt.wantCode, w.Code)
		assert.Equal(hs.T(), tt.wantErr.Error(), res.Err)
	}
}

func (hs *HandlerTestSuite) TestGetUserFriends() {
	u := "friend"
	_, _ = hs.svc.RegisterNewUser(registerUserRequest{u, "password", "uf@uf.co"})
	uid := string(hs.userID)
	_ = hs.svc.CreateRelationshipFor(ID(uid), u)

	tests := []struct {
		username           string
		wantCode, wantFLen int
	}{
		{username: "  ", wantCode: http.StatusBadRequest},
		{username: "nonexistent", wantCode: http.StatusNotFound},
		{username: hs.req.Username, wantCode: http.StatusOK, wantFLen: 1},
	}

	for _, tt := range tests {
		url := fmt.Sprintf("/v1/users/%s/friends", tt.username)
		r, _ := http.NewRequest(http.MethodGet, url, nil)

		router := httprouter.New()
		router.Handler(http.MethodGet, "/v1/users/:username/friends", GetUserFriendsHandler(hs.svc))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		var res []UserInfo

		_ = json.NewDecoder(w.Body).Decode(&res)

		assert.Equal(hs.T(), tt.wantCode, w.Code)
		assert.Equal(hs.T(), tt.wantFLen, len(res))
	}

	_ = hs.svc.RemoveRelationshipFor(ID(uid), u)
}

func (hs *HandlerTestSuite) TestGetUserFollowers() {
	u := "follower"
	id, _ := hs.svc.RegisterNewUser(registerUserRequest{u, "password", "usf@uf.co"})
	_ = hs.svc.CreateRelationshipFor(id, hs.req.Username)

	tests := []struct {
		username           string
		wantCode, wantFLen int
	}{
		{username: "  ", wantCode: http.StatusBadRequest},
		{username: "nonexistent", wantCode: http.StatusNotFound},
		{username: hs.req.Username, wantCode: http.StatusOK, wantFLen: 1},
	}

	for _, tt := range tests {
		url := fmt.Sprintf("/v1/users/%s/followers", tt.username)
		r, _ := http.NewRequest(http.MethodGet, url, nil)

		router := httprouter.New()
		router.Handler(http.MethodGet, "/v1/users/:username/followers", GetUserFollowersHandler(hs.svc))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		var res []UserInfo

		_ = json.NewDecoder(w.Body).Decode(&res)

		assert.Equal(hs.T(), tt.wantCode, w.Code)
		assert.Equal(hs.T(), tt.wantFLen, len(res))
	}

	_ = hs.svc.RemoveRelationshipFor(id, u)
}

var invalidToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU"

func TestRequireAuthMiddleware(t *testing.T) {
	validToken, _ := getJWTToken("randomid")

	tests := []struct {
		authHeader string
		wantCode   int
		wantID     string
	}{
		{wantCode: http.StatusUnauthorized},
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
	validToken, _ := getJWTToken(string(hs.userID))

	var called bool
	f := func(w http.ResponseWriter, r *http.Request) {
		called = true
	}

	ls := LastSeenMiddleware(http.HandlerFunc(f), hs.svc)
	r, _ := http.NewRequest("", "/doesnt-matter", nil)

	tests := []struct {
		id, token                       string
		withCtx, withCtxHeader          bool
		wantUpdatedLastSeen, wantCalled bool
	}{
		{wantCalled: true},
		{token: invalidToken, withCtxHeader: true, wantCalled: true},
		{token: validToken, withCtxHeader: true, wantUpdatedLastSeen: true, wantCalled: true},
		{id: string(hs.userID), withCtx: true, wantUpdatedLastSeen: true, wantCalled: true},
	}

	for _, tt := range tests {
		if tt.withCtx {
			r = r.WithContext(context.WithValue(r.Context(), idKey, tt.id))
		}

		if tt.withCtxHeader {
			r.Header.Set("Authorization", "Bearer "+validToken)
		}

		w := httptest.NewRecorder()
		router := httprouter.New()
		router.Handler(http.MethodGet, "/doesnt-matter", ls)
		router.ServeHTTP(w, r)

		assert.IsType(hs.T(), new(http.Handler), &ls)
		assert.Equal(hs.T(), http.StatusOK, w.Code)
		assert.Equal(hs.T(), tt.wantCalled, called)

		if tt.wantUpdatedLastSeen {
			p, _ := hs.svc.GetProfile(hs.req.Username)
			assert.Equal(hs.T(), tt.wantUpdatedLastSeen, p.LastSeen.After(now))
		}
	}
}

func setIDInRequestContext(r *http.Request, id string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), idKey, id))
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
