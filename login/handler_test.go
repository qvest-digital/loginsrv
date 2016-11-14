package login

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
			&Config{Backends: BackendOptions{map[string]string{"provider": "simple", "bob": "secret"}}},
			1,
			false,
		},
		// error cases
		{
			&Config{},
			0,
			true,
		},
		{
			&Config{Backends: BackendOptions{map[string]string{"foo": ""}}},
			1,
			true,
		},
		{
			&Config{Backends: BackendOptions{map[string]string{"provider": "simpleFoo", "bob": "secret"}}},
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

func TestHandler_404(t *testing.T) {
	recorder := call(req("GET", "/foo", ""))
	assert.Equal(t, recorder.Code, 404)
}

func TestHandler_LoginForm(t *testing.T) {
	recorder := call(req("GET", "/context/login", ""))
	assert.Equal(t, recorder.Code, 200)
	assert.Contains(t, recorder.Body.String(), "form")
	assert.Contains(t, recorder.Body.String(), `method="POST"`)
	assert.Contains(t, recorder.Body.String(), `action="/context/login"`)
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
	assert.True(t, recorder.Body.Len() > 30)

	// TODO: verify the jwt token

	// wrong credentials
	recorder = call(req("POST", "/context/login", `{"username": "bob", "password": "FOOOBAR"}`, TypeJson, AcceptJwt))
	assert.Equal(t, 403, recorder.Code)
	assert.Equal(t, "Wrong credentials", recorder.Body.String())
}

func TestHandler_LoginWeb(t *testing.T) {
	// redirectSuccess
	recorder := call(req("POST", "/context/login", "username=bob&password=secret", TypeForm, AcceptHtml))
	assert.Equal(t, 303, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Set-Cookie"), "jwt_token=")
	assert.Equal(t, "/", recorder.Header().Get("Location"))

	// show the login after error
	recorder = call(req("POST", "/context/login", "username=bob&password=FOOBAR", TypeForm, AcceptHtml))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "form")
	assert.Contains(t, recorder.Body.String(), `method="POST"`)
	assert.Contains(t, recorder.Body.String(), `action="/context/login"`)
	assert.Equal(t, recorder.Header().Get("Set-Cookie"), "")
}

func testHandler() *Handler {
	return &Handler{
		backends: []Backend{
			NewSimpleBackend(map[string]string{"bob": "secret"}),
		},
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
