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
		expectError  bool
	}{
		{
			&Config{Backends: Options{"simple": map[string]string{"bob": "secret"}}},
			1,
			false,
		},
		// error cases
		{
			// init error because no users are provided
			&Config{Backends: Options{"simple": map[string]string{}}},
			1,
			true,
		},
		{
			&Config{},
			0,
			true,
		},
		{
			&Config{Backends: Options{"simpleFoo": map[string]string{"bob": "secret"}}},
			1,
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
				assert.Equal(t, 1, len(h.backends))
			}
		})
	}
}

func TestHandler_LoginForm(t *testing.T) {
	recorder := call(req("GET", "/context/login", ""))
	assert.Equal(t, recorder.Code, 200)
	assert.Contains(t, recorder.Body.String(), `class="container`)
	assert.Equal(t, "no-cache, no-store, must-revalidate", recorder.Header().Get("Cache-Control"))
}

func TestHandler_HEAD(t *testing.T) {
	recorder := call(req("HEAD", "/context/login", ""))
	assert.Equal(t, recorder.Code, 400)
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
	return &Handler{
		backends: []Backend{
			NewSimpleBackend(map[string]string{"bob": "secret"}),
		},
		oauth:  oauth2.NewManager(),
		config: DefaultConfig(),
	}
}

func testHandlerWithError() *Handler {
	return &Handler{
		backends: []Backend{
			errorTestBackend("test error"),
		},
		oauth:  oauth2.NewManager(),
		config: DefaultConfig(),
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
