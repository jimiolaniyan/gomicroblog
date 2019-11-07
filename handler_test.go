package blog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/jimiolaniyan/gomicroblog/auth"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/stretchr/testify/require"

	"github.com/julienschmidt/httprouter"

	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/assert"
)

type HandlerTestSuite struct {
	suite.Suite
	userID      ID
	svc         Service
	username    string
	user        *User
	users       Repository
	containerID string
	client      *mongo.Client
}

func (hs *HandlerTestSuite) SetupSuite() {
	var users Repository
	var posts PostRepository

	if !testing.Short() {
		containerID, err := RunDockerContainer("mongo:latest")
		require.NoError(hs.T(), err)

		hs.containerID = containerID

		log.Println("Container id:", containerID)

		ip, _ := GetContainerIP(containerID)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+ip+":27017"))
		require.NoError(hs.T(), err)

		err = client.Ping(ctx, nil)
		require.NoError(hs.T(), err)

		hs.client = client

		u := client.Database("testing").Collection("users")
		p := client.Database("testing").Collection("posts")
		users = NewMongoUserRepository(u)
		posts = NewMongoPostRepository(p)
	} else {
		users = NewUserRepository()
		posts = NewPostRepository()
	}

	hs.users = users
	hs.svc = NewService(users, posts)

	id := nextID()
	username := "user"
	hs.svc.CreateProfile(string(id), username, "a@b.con")

	hs.username = username
	hs.userID = id

	u, _ := users.FindByID(id)
	u.Password, _ = hashPassword("password")
	hs.user = u
	_ = hs.users.Update(u)
}

func (hs *HandlerTestSuite) TearDownSuite() {
	if !testing.Short() {
		_ = hs.client.Disconnect(context.Background())

		// kill docker container
		_ = exec.Command("docker", "kill", hs.containerID).Run()
	}
}

var errNil = errors.New("")

func (hs *HandlerTestSuite) TestAccountCreatedHandler() {
	req := `{"username": "u", "email": "e@m.com", "password": "password"}`
	r, _ := http.NewRequest(http.MethodPost, "/auth/v1/accounts", strings.NewReader(req))

	w := httptest.NewRecorder()
	handler := http.NewServeMux()
	svc := auth.NewService(auth.NewAccountRepository(), NewAccountCreatedHandler(hs.svc))
	handler.Handle("/auth/v1/accounts", auth.RegisterAccountHandler(svc))
	handler.ServeHTTP(w, r)

	var res struct {
		ID ID `json:"id"`
	}

	_ = json.NewDecoder(w.Body).Decode(&res)

	u, _ := hs.users.FindByID(res.ID)
	assert.NotNil(hs.T(), u)
	assert.Equal(hs.T(), u.ID, res.ID)
}

func (hs *HandlerTestSuite) TestCreatePostHandler() {
	uid := string(hs.userID)
	b := `{"body": ""}`
	body := `{"body": "i love my wife :)"}`

	tests := []struct {
		req, userID     string
		wantCode        int
		wantErr         error
		wantID, withCtx bool
		wantLoc         string
	}{
		{req: `invalid request`, wantCode: http.StatusBadRequest, wantErr: errNil, withCtx: true},
		{req: `{}`, wantCode: http.StatusInternalServerError, wantErr: ErrEmptyContext},
		{req: b, wantCode: http.StatusUnauthorized, wantErr: ErrInvalidID, withCtx: true},
		{req: b, userID: "puoiwoerigp", wantCode: http.StatusUnauthorized, wantErr: ErrInvalidID, withCtx: true},
		{req: b, userID: uid, wantCode: http.StatusUnprocessableEntity, wantErr: ErrEmptyBody, withCtx: true},
		{req: body, userID: uid, wantCode: http.StatusCreated, wantErr: errNil, wantID: true, wantLoc: "/v1/posts/", withCtx: true},
	}

	for _, tt := range tests {
		r, _ := http.NewRequest(http.MethodPost, "/v1/posts", strings.NewReader(tt.req))

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.Handle("/v1/posts", CreatePostHandler(hs.svc))

		if tt.withCtx {
			r = r.WithContext(context.WithValue(r.Context(), idKey, tt.userID))
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
		assert.True(hs.T(), strings.HasPrefix(w.Header().Get("Location"), tt.wantLoc))
	}

}

func (hs *HandlerTestSuite) TestGetProfileHandler() {
	u := "postu"
	host := "http://localhost:8080"
	finalURL := fmt.Sprintf("%s/v1/users/%s", host, u)

	user := DuplicateUser(hs.users, *hs.user, u)
	_, _ = hs.svc.CreatePost(hs.userID, "post")
	_, _ = hs.svc.CreatePost(user.ID, "post")

	tests := []struct {
		username              string
		wantCode              int
		wantErr               error
		wantID                bool
		wantUsername, wantURL string
		wantPosLen            int
	}{
		{username: "  ", wantCode: http.StatusBadRequest, wantErr: errNil, wantUsername: ""},
		{username: "nonexistent", wantCode: http.StatusNotFound, wantErr: ErrNotFound, wantUsername: ""},
		{username: u, wantCode: http.StatusOK, wantErr: errNil, wantID: true, wantUsername: u, wantURL: finalURL, wantPosLen: 1},
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
		assert.Equal(hs.T(), tt.wantPosLen, len(res.Profile.Posts))
		assert.Equal(hs.T(), tt.wantURL, res.URL)
	}

}

func (hs *HandlerTestSuite) TestEditProfileHandler() {
	// duplicate user to avoid conflicts with user in the suite
	username := "tempUser"
	user := DuplicateUser(hs.users, *hs.user, username)
	sid := string(user.ID)

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
		{req: `invalid request`, wantCode: http.StatusBadRequest, wantErr: errNil},
		{req: `{}`, wantCode: http.StatusInternalServerError, wantErr: ErrEmptyContext},
		{req: `{}`, id: "invalid", wantCode: http.StatusUnauthorized, withCtx: true, wantErr: ErrInvalidID},
		{req: `{}`, id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: errNil},
		{req: `{"username": ""}`, id: sid, wantCode: S422, withCtx: true, wantErr: ErrInvalidUsername},
		{req: fmt.Sprintf(`{"username": "%s"}`, username), id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: errNil},
		{req: fmt.Sprintf(`{"username": "%s"}`, hs.username), id: sid, wantCode: http.StatusConflict, withCtx: true, wantErr: ErrExistingUsername},
		{req: `{"username": "newName"}`, id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: errNil, wantUsername: "newName", reset: true},
		{req: `{"bio": ""}`, id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: errNil},
		{req: fmt.Sprintf(`{"bio": "%s"}`, longBio), id: sid, wantCode: S422, withCtx: true, wantErr: ErrBioTooLong},
		{req: `{"bio": "Bios"}`, id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: errNil, wantBio: "Bios"},
		{req: `{"username": "newU", "bio": "Be nice"}`, id: sid, wantCode: http.StatusOK, withCtx: true, wantErr: errNil, wantBio: "Be nice", wantUsername: "newU"},
	}

	for _, tt := range tests {
		r, _ := http.NewRequest(http.MethodPatch, "/v1/users", strings.NewReader(tt.req))

		u := username
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
			assert.Equal(hs.T(), username, res2.Profile.Username)
		}

		//reset the username and bio
		if tt.reset {
			body := strings.NewReader(fmt.Sprintf(`{"username": "%s", "bio": ""}`, username))
			r3, _ := http.NewRequest(http.MethodPatch, "/v1/users", body)
			r3 = r3.WithContext(context.WithValue(r3.Context(), idKey, tt.id))
			router.ServeHTTP(w, r3)
		}
	}
}

func (hs *HandlerTestSuite) TestCreateRelationshipHandler() {
	uid := string(hs.userID)

	u := "followUser"
	DuplicateUser(hs.users, *hs.user, u)

	tests := []struct {
		username string
		withCtx  bool
		wantCode int
		wantErr  error
	}{
		{username: "  ", wantCode: http.StatusBadRequest, wantErr: errNil},
		{username: "nonexistent", wantCode: http.StatusInternalServerError, wantErr: ErrEmptyContext},
		{username: "nonexistent", withCtx: true, wantCode: http.StatusNotFound, wantErr: ErrNotFound},
		{username: hs.username, withCtx: true, wantCode: http.StatusForbidden, wantErr: ErrCantFollowSelf},
		{username: u, withCtx: true, wantCode: http.StatusNoContent, wantErr: errNil},
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
	DuplicateUser(hs.users, *hs.user, u)
	_ = hs.svc.CreateRelationshipFor(ID(uid), u)

	tests := []struct {
		username string
		withCtx  bool
		wantCode int
		wantErr  error
	}{
		{username: "  ", wantCode: http.StatusBadRequest, wantErr: errNil},
		{username: "nonexistent", wantCode: http.StatusInternalServerError, wantErr: ErrEmptyContext},
		{username: "nonexistent", withCtx: true, wantCode: http.StatusNotFound, wantErr: ErrNotFound},
		{username: hs.username, withCtx: true, wantCode: http.StatusForbidden, wantErr: ErrCantUnFollowSelf},
		{username: u, withCtx: true, wantCode: http.StatusNoContent, wantErr: errNil},
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

func (hs *HandlerTestSuite) TestGetRelationships() {
	u := "follower"
	user := DuplicateUser(hs.users, *hs.user, u)
	_ = hs.svc.CreateRelationshipFor(user.ID, hs.username)
	_ = hs.svc.CreateRelationshipFor(hs.userID, u)

	tests := []struct {
		username           string
		wantCode, wantFLen int
	}{

		{username: "  ", wantCode: http.StatusBadRequest},
		{username: "nonexistent", wantCode: http.StatusNotFound},
		{username: hs.username, wantCode: http.StatusOK, wantFLen: 1},
	}

	for _, tt := range tests {
		url1 := fmt.Sprintf("/v1/users/%s/friends", tt.username)
		url2 := fmt.Sprintf("/v1/users/%s/followers", tt.username)

		r1, _ := http.NewRequest(http.MethodGet, url1, nil)
		r2, _ := http.NewRequest(http.MethodGet, url2, nil)

		router := httprouter.New()
		router.Handler(http.MethodGet, "/v1/users/:username/friends", GetUserFriendsHandler(hs.svc))
		router.Handler(http.MethodGet, "/v1/users/:username/followers", GetUserFollowersHandler(hs.svc))

		w1 := httptest.NewRecorder()
		w2 := httptest.NewRecorder()

		router.ServeHTTP(w1, r1)
		router.ServeHTTP(w2, r2)

		var res1 []UserInfo
		var res2 []UserInfo

		_ = json.NewDecoder(w1.Body).Decode(&res1)
		_ = json.NewDecoder(w2.Body).Decode(&res2)

		assert.Equal(hs.T(), tt.wantCode, w1.Code)
		assert.Equal(hs.T(), tt.wantCode, w2.Code)
		assert.Equal(hs.T(), tt.wantFLen, len(res1))
		assert.Equal(hs.T(), tt.wantFLen, len(res2))
	}
}

var invalidToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU"

func (hs *HandlerTestSuite) TestRequireAuthMiddleware() {
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

		assert.IsType(hs.T(), new(http.Handler), &h)
		assert.Equal(hs.T(), tt.wantID, id)
		assert.Equal(hs.T(), tt.wantCode, w.Code)
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
			p, _ := hs.svc.GetProfile(hs.username)
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
