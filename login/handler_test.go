package login

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/model"
	"github.com/tarent/loginsrv/oauth2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const TypeJson = "Content-Type: application/json"
const TypeForm = "Content-Type: application/x-www-form-urlencoded"
const AcceptHtml = "Accept: text/html"
const AcceptJwt = "Accept: application/jwt"

func TestHandler_NewFromConfig(t *testing.T) {

	testCases := []struct {
		config       *Config
		backendCount int
		oauthCount   int
		expectError  bool
	}{
		{
			&Config{
				Backends: Options{
					"simple": map[string]string{"bob": "secret"},
				},
				Oauth: Options{
					"github": map[string]string{"client_id": "xxx", "client_secret": "YYY"},
				},
			},
			1,
			1,
			false,
		},
		{
			&Config{Backends: Options{"simple": map[string]string{"bob": "secret"}}},
			1,
			0,
			false,
		},
		// error cases
		{
			// init error because no users are provided
			&Config{Backends: Options{"simple": map[string]string{}}},
			1,
			0,
			true,
		},
		{
			&Config{
				Oauth: Options{
					"FOOO": map[string]string{"client_id": "xxx", "client_secret": "YYY"},
				},
			},
			0,
			0,
			true,
		},
		{
			&Config{},
			0,
			0,
			true,
		},
		{
			&Config{Backends: Options{"simpleFoo": map[string]string{"bob": "secret"}}},
			1,
			0,
			true,
		},
	}
	for i, test := range testCases {
		t.Run(fmt.Sprintf("test %v", i), func(t *testing.T) {
			h, err := NewHandler(test.config)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.backendCount, len(h.backends))
				assert.Equal(t, test.oauthCount, len(h.oauth.(*oauth2.Manager).GetConfigs()))
			}
		})
	}
}

func TestHandler_LoginForm(t *testing.T) {
	recorder := call(req("GET", "/context/login", ""))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `class="container`)
	assert.Equal(t, "no-cache, no-store, must-revalidate", recorder.Header().Get("Cache-Control"))
}

func TestHandler_HEAD(t *testing.T) {
	recorder := call(req("HEAD", "/context/login", ""))
	assert.Equal(t, 400, recorder.Code)
}

func TestHandler_404(t *testing.T) {
	recorder := call(req("GET", "/context/", ""))
	assert.Equal(t, 404, recorder.Code)

	recorder = call(req("GET", "/", ""))
	assert.Equal(t, 404, recorder.Code)

	assert.Equal(t, "Not Found: The requested page does not exist", recorder.Body.String())
}

func TestHandler_LoginJson(t *testing.T) {
	// success
	recorder := call(req("POST", "/context/login", `{"username": "bob", "password": "secret"}`, TypeJson, AcceptJwt))
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, recorder.Header().Get("Content-Type"), "application/jwt")

	// verify the token
	claims, err := tokenAsMap(recorder.Body.String())
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"sub": "bob"}, claims)

	// wrong credentials
	recorder = call(req("POST", "/context/login", `{"username": "bob", "password": "FOOOBAR"}`, TypeJson, AcceptJwt))
	assert.Equal(t, 403, recorder.Code)
	assert.Equal(t, "Wrong credentials", recorder.Body.String())
}

func TestHandler_HandleOauth(t *testing.T) {
	managerMock := &oauth2ManagerMock{
		_GetConfigFromRequest: func(r *http.Request) (oauth2.Config, error) {
			return oauth2.Config{}, nil
		},
	}
	handler := &Handler{
		oauth:  managerMock,
		config: DefaultConfig(),
	}

	// test start flow redirect
	managerMock._Handle = func(w http.ResponseWriter, r *http.Request) (
		startedFlow bool,
		authenticated bool,
		userInfo model.UserInfo,
		err error) {
		w.Header().Set("Location", "http://example.com")
		w.WriteHeader(303)
		return true, false, model.UserInfo{}, nil
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req("GET", "/login/github", ""))
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "http://example.com", recorder.Header().Get("Location"))

	// test authentication
	managerMock._Handle = func(w http.ResponseWriter, r *http.Request) (
		startedFlow bool,
		authenticated bool,
		userInfo model.UserInfo,
		err error) {
		return false, true, model.UserInfo{Sub: "marvin"}, nil
	}
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req("GET", "/login/github", ""))
	assert.Equal(t, 200, recorder.Code)
	token, err := tokenAsMap(recorder.Body.String())
	assert.NoError(t, err)
	assert.Equal(t, "marvin", token["sub"])

	// test error in oauth
	managerMock._Handle = func(w http.ResponseWriter, r *http.Request) (
		startedFlow bool,
		authenticated bool,
		userInfo model.UserInfo,
		err error) {
		return false, false, model.UserInfo{}, errors.New("some error")
	}
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req("GET", "/login/github", ""))
	assert.Equal(t, 500, recorder.Code)

	// test failure if no oauth action would be taken, because the url parameters where
	// missing an action parts
	managerMock._Handle = func(w http.ResponseWriter, r *http.Request) (
		startedFlow bool,
		authenticated bool,
		userInfo model.UserInfo,
		err error) {
		return false, false, model.UserInfo{}, nil
	}
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req("GET", "/login/github", ""))
	assert.Equal(t, 403, recorder.Code)
}

func TestHandler_LoginWeb(t *testing.T) {
	// redirectSuccess
	recorder := call(req("POST", "/context/login", "username=bob&password=secret", TypeForm, AcceptHtml))
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "/", recorder.Header().Get("Location"))

	// verify the token from the cookie
	assert.Contains(t, recorder.Header().Get("Set-Cookie"), "jwt_token=")
	headerParts := strings.SplitN(recorder.Header().Get("Set-Cookie"), "=", 2)
	assert.Equal(t, 2, len(headerParts))
	assert.Equal(t, headerParts[0], "jwt_token")
	claims, err := tokenAsMap(strings.SplitN(headerParts[1], ";", 2)[0])
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"sub": "bob"}, claims)
	assert.Contains(t, headerParts[1]+";", "Path=/;")

	// show the login form again after authentication failed
	recorder = call(req("POST", "/context/login", "username=bob&password=FOOBAR", TypeForm, AcceptHtml))
	assert.Equal(t, 403, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `class="container"`)
	assert.Equal(t, recorder.Header().Get("Set-Cookie"), "")
}

func TestHandler_Logout(t *testing.T) {
	// DELETE
	recorder := call(req("DELETE", "/context/login", ""))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Set-Cookie"), "jwt_token=delete; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT;")

	// GET  + param
	recorder = call(req("GET", "/context/login?logout=true", ""))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Set-Cookie"), "jwt_token=delete; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT;")

	// POST + param
	recorder = call(req("POST", "/context/login", "logout=true", TypeForm))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Set-Cookie"), "jwt_token=delete; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT;")

	assert.Equal(t, "no-cache, no-store, must-revalidate", recorder.Header().Get("Cache-Control"))
}

func TestHandler_LoginError(t *testing.T) {
	h := testHandlerWithError()

	// backend returning an error with result type == jwt
	request := req("POST", "/context/login", `{"username": "bob", "password": "secret"}`, TypeJson, AcceptJwt)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, request)

	assert.Equal(t, 500, recorder.Code)
	assert.Equal(t, recorder.Header().Get("Content-Type"), "text/plain")
	assert.Equal(t, recorder.Body.String(), "Internal Server Error")

	// backend returning an error with result type == html
	request = req("POST", "/context/login", `{"username": "bob", "password": "secret"}`, TypeJson, AcceptHtml)
	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, request)

	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, recorder.Body.String(), `class="container"`)
	assert.Contains(t, recorder.Body.String(), "Internal Error")
}

func TestHandler_getToken_Valid(t *testing.T) {
	h := testHandler()
	input := model.UserInfo{Sub: "marvin"}
	token, err := h.createToken(input)
	assert.NoError(t, err)
	r := &http.Request{
		Header: http.Header{"Cookie": {h.config.CookieName + "=" + token + ";"}},
	}
	userInfo, valid := h.getToken(r)
	assert.True(t, valid)
	assert.Equal(t, input, userInfo)
}

func TestHandler_getToken_InvalidSecret(t *testing.T) {
	h := testHandler()
	input := model.UserInfo{Sub: "marvin"}
	token, err := h.createToken(input)
	assert.NoError(t, err)
	r := &http.Request{
		Header: http.Header{"Cookie": {h.config.CookieName + "=" + token + ";"}},
	}

	// modify secret
	h.config.JwtSecret = "foobar"

	_, valid := h.getToken(r)
	assert.False(t, valid)
}

func TestHandler_getToken_InvalidToken(t *testing.T) {
	h := testHandler()
	r := &http.Request{
		Header: http.Header{"Cookie": {h.config.CookieName + "=asdcsadcsadc"}},
	}

	_, valid := h.getToken(r)
	assert.False(t, valid)
}

func TestHandler_getToken_InvalidNoToken(t *testing.T) {
	h := testHandler()
	_, valid := h.getToken(&http.Request{})
	assert.False(t, valid)
}

func testHandler() *Handler {
	cfg := DefaultConfig()
	cfg.LoginPath = "/context/login"
	return &Handler{
		backends: []Backend{
			NewSimpleBackend(map[string]string{"bob": "secret"}),
		},
		oauth:  oauth2.NewManager(),
		config: cfg,
	}
}

func testHandlerWithError() *Handler {
	cfg := DefaultConfig()
	cfg.LoginPath = "/context/login"
	return &Handler{
		backends: []Backend{
			errorTestBackend("test error"),
		},
		oauth:  oauth2.NewManager(),
		config: cfg,
	}
}

func call(req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	h := testHandler()
	h.ServeHTTP(recorder, req)
	return recorder
}

func req(method string, url string, body string, header ...string) *http.Request {
	r, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	for _, h := range header {
		pair := strings.SplitN(h, ": ", 2)
		r.Header.Add(pair[0], pair[1])
	}
	return r
}

func tokenAsMap(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return []byte(DefaultConfig().JwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return map[string]interface{}(claims), nil
	} else {
		return nil, errors.New("token not valid")
	}
}

type errorTestBackend string

func (h errorTestBackend) Authenticate(username, password string) (bool, model.UserInfo, error) {
	return false, model.UserInfo{}, errors.New(string(h))
}

type oauth2ManagerMock struct {
	_Handle func(w http.ResponseWriter, r *http.Request) (
		startedFlow bool,
		authenticated bool,
		userInfo model.UserInfo,
		err error)
	_AddConfig            func(providerName string, opts map[string]string) error
	_GetConfigFromRequest func(r *http.Request) (oauth2.Config, error)
}

func (m *oauth2ManagerMock) Handle(w http.ResponseWriter, r *http.Request) (
	startedFlow bool,
	authenticated bool,
	userInfo model.UserInfo,
	err error) {
	return m._Handle(w, r)
}
func (m *oauth2ManagerMock) AddConfig(providerName string, opts map[string]string) error {
	return m._AddConfig(providerName, opts)
}
func (m *oauth2ManagerMock) GetConfigFromRequest(r *http.Request) (oauth2.Config, error) {
	return m._GetConfigFromRequest(r)
}
