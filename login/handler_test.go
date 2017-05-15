package login

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/model"
	"github.com/tarent/loginsrv/oauth2"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const TypeJSON = "Content-Type: application/json"
const TypeForm = "Content-Type: application/x-www-form-urlencoded"
const AcceptHTML = "Accept: text/html"
const AcceptJwt = "Accept: application/jwt"

func testConfig() *Config {
	testConfig := DefaultConfig()
	testConfig.LoginPath = "/context/login"
	testConfig.CookieDomain = "example.com"
	testConfig.CookieExpiry = 23 * time.Hour
	return testConfig
}

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
					"simple": {"bob": "secret"},
				},
				Oauth: Options{
					"github": {"client_id": "xxx", "client_secret": "YYY"},
				},
			},
			1,
			1,
			false,
		},
		{
			&Config{Backends: Options{"simple": {"bob": "secret"}}},
			1,
			0,
			false,
		},
		// error cases
		{
			// init error because no users are provided
			&Config{Backends: Options{"simple": {}}},
			1,
			0,
			true,
		},
		{
			&Config{
				Oauth: Options{
					"FOOO": {"client_id": "xxx", "client_secret": "YYY"},
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
			&Config{Backends: Options{"simpleFoo": {"bob": "secret"}}},
			1,
			0,
			true,
		},
	}
	for i, test := range testCases {
		t.Run(fmt.Sprintf("test %v", i), func(t *testing.T) {
			h, err := NewHandler(test.config)
			if test.expectError {
				Error(t, err)
			} else {
				NoError(t, err)
				Equal(t, test.backendCount, len(h.backends))
				Equal(t, test.oauthCount, len(h.oauth.(*oauth2.Manager).GetConfigs()))
			}
		})
	}
}

func TestHandler_LoginForm(t *testing.T) {
	recorder := call(req("GET", "/context/login", ""))
	Equal(t, 200, recorder.Code)
	Contains(t, recorder.Body.String(), `class="container`)
	Equal(t, "no-cache, no-store, must-revalidate", recorder.Header().Get("Cache-Control"))
}

func TestHandler_HEAD(t *testing.T) {
	recorder := call(req("HEAD", "/context/login", ""))
	Equal(t, 400, recorder.Code)
}

func TestHandler_404(t *testing.T) {
	recorder := call(req("GET", "/context/", ""))
	Equal(t, 404, recorder.Code)

	recorder = call(req("GET", "/", ""))
	Equal(t, 404, recorder.Code)

	Equal(t, "Not Found: The requested page does not exist", recorder.Body.String())
}

func TestHandler_LoginJson(t *testing.T) {
	// success
	recorder := call(req("POST", "/context/login", `{"username": "bob", "password": "secret"}`, TypeJSON, AcceptJwt))
	Equal(t, 200, recorder.Code)
	Equal(t, recorder.Header().Get("Content-Type"), "application/jwt")

	// verify the token
	claims, err := tokenAsMap(recorder.Body.String())
	NoError(t, err)
	Equal(t, "bob", claims["sub"])
	InDelta(t, time.Now().Add(DefaultConfig().JwtExpiry).Unix(), claims["exp"], 2)

	// wrong credentials
	recorder = call(req("POST", "/context/login", `{"username": "bob", "password": "FOOOBAR"}`, TypeJSON, AcceptJwt))
	Equal(t, 403, recorder.Code)
	Equal(t, "Wrong credentials", recorder.Body.String())
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
	Equal(t, 303, recorder.Code)
	Equal(t, "http://example.com", recorder.Header().Get("Location"))

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
	Equal(t, 200, recorder.Code)
	token, err := tokenAsMap(recorder.Body.String())
	NoError(t, err)
	Equal(t, "marvin", token["sub"])

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
	Equal(t, 500, recorder.Code)

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
	Equal(t, 403, recorder.Code)
}

func TestHandler_LoginWeb(t *testing.T) {
	// redirectSuccess
	recorder := call(req("POST", "/context/login", "username=bob&password=secret", TypeForm, AcceptHTML))
	Equal(t, 303, recorder.Code)
	Equal(t, "/", recorder.Header().Get("Location"))

	// verify the token from the cookie
	setCookieList := readSetCookies(recorder.Header())
	Equal(t, 1, len(setCookieList))

	cookie := setCookieList[0]
	Equal(t, "jwt_token", cookie.Name)
	Equal(t, "/", cookie.Path)
	Equal(t, "example.com", cookie.Domain)
	InDelta(t, time.Now().Add(testConfig().CookieExpiry).Unix(), cookie.Expires.Unix(), 2)
	True(t, cookie.HttpOnly)

	// check the token content
	claims, err := tokenAsMap(cookie.Value)
	NoError(t, err)
	Equal(t, "bob", claims["sub"])
	InDelta(t, time.Now().Add(DefaultConfig().JwtExpiry).Unix(), claims["exp"], 2)

	// show the login form again after authentication failed
	recorder = call(req("POST", "/context/login", "username=bob&password=FOOBAR", TypeForm, AcceptHTML))
	Equal(t, 403, recorder.Code)
	Contains(t, recorder.Body.String(), `class="container"`)
	Equal(t, recorder.Header().Get("Set-Cookie"), "")
}

func TestHandler_Logout(t *testing.T) {
	// DELETE
	recorder := call(req("DELETE", "/context/login", ""))
	Equal(t, 200, recorder.Code)
	checkDeleteCookei(t, recorder.Header())

	// GET  + param
	recorder = call(req("GET", "/context/login?logout=true", ""))
	Equal(t, 200, recorder.Code)
	checkDeleteCookei(t, recorder.Header())

	// POST + param
	recorder = call(req("POST", "/context/login", "logout=true", TypeForm))
	Equal(t, 200, recorder.Code)
	checkDeleteCookei(t, recorder.Header())

	Equal(t, "no-cache, no-store, must-revalidate", recorder.Header().Get("Cache-Control"))
}

func checkDeleteCookei(t *testing.T, h http.Header) {
	setCookieList := readSetCookies(h)
	Equal(t, 1, len(setCookieList))
	cookie := setCookieList[0]

	Equal(t, "jwt_token", cookie.Name)
	Equal(t, "/", cookie.Path)
	Equal(t, "example.com", cookie.Domain)
	Equal(t, int64(0), cookie.Expires.Unix())
}

func TestHandler_CustomLogoutURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LogoutURL = "http://example.com"
	h := &Handler{
		oauth:  oauth2.NewManager(),
		config: cfg,
	}

	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req("DELETE", "/login", ""))
	Contains(t, recorder.Header().Get("Set-Cookie"), "jwt_token=delete; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT;")
	Equal(t, 303, recorder.Code)
	Equal(t, "http://example.com", recorder.Header().Get("Location"))
}

func TestHandler_LoginError(t *testing.T) {
	h := testHandlerWithError()

	// backend returning an error with result type == jwt
	request := req("POST", "/context/login", `{"username": "bob", "password": "secret"}`, TypeJSON, AcceptJwt)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, request)

	Equal(t, 500, recorder.Code)
	Equal(t, recorder.Header().Get("Content-Type"), "text/plain")
	Equal(t, recorder.Body.String(), "Internal Server Error")

	// backend returning an error with result type == html
	request = req("POST", "/context/login", `{"username": "bob", "password": "secret"}`, TypeJSON, AcceptHTML)
	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, request)

	Equal(t, 500, recorder.Code)
	Contains(t, recorder.Header().Get("Content-Type"), "text/html")
	Contains(t, recorder.Body.String(), `class="container"`)
	Contains(t, recorder.Body.String(), "Internal Error")
}

func TestHandler_getToken_Valid(t *testing.T) {
	h := testHandler()
	input := model.UserInfo{Sub: "marvin", Expiry: time.Now().Add(time.Second).Unix()}
	token, err := h.createToken(input)
	NoError(t, err)
	r := &http.Request{
		Header: http.Header{"Cookie": {h.config.CookieName + "=" + token + ";"}},
	}
	userInfo, valid := h.getToken(r)
	True(t, valid)
	Equal(t, input, userInfo)
}

func TestHandler_getToken_InvalidSecret(t *testing.T) {
	h := testHandler()
	input := model.UserInfo{Sub: "marvin"}
	token, err := h.createToken(input)
	NoError(t, err)
	r := &http.Request{
		Header: http.Header{"Cookie": {h.config.CookieName + "=" + token + ";"}},
	}

	// modify secret
	h.config.JwtSecret = "foobar"

	_, valid := h.getToken(r)
	False(t, valid)
}

func TestHandler_getToken_InvalidToken(t *testing.T) {
	h := testHandler()
	r := &http.Request{
		Header: http.Header{"Cookie": {h.config.CookieName + "=asdcsadcsadc"}},
	}

	_, valid := h.getToken(r)
	False(t, valid)
}

func TestHandler_getToken_InvalidNoToken(t *testing.T) {
	h := testHandler()
	_, valid := h.getToken(&http.Request{})
	False(t, valid)
}

func testHandler() *Handler {
	return &Handler{
		backends: []Backend{
			NewSimpleBackend(map[string]string{"bob": "secret"}),
		},
		oauth:  oauth2.NewManager(),
		config: testConfig(),
	}
}

func testHandlerWithError() *Handler {
	return &Handler{
		backends: []Backend{
			errorTestBackend("test error"),
		},
		oauth:  oauth2.NewManager(),
		config: testConfig(),
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
	}

	return nil, errors.New("token not valid")
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

// copied from golang: net/http/cookie.go
// with simple some simplification fro edge cases
// readSetCookies parses all "Set-Cookie" values from
// the header h and returns the successfully parsed Cookies.
func readSetCookies(h http.Header) []*http.Cookie {
	cookieCount := len(h["Set-Cookie"])
	if cookieCount == 0 {
		return []*http.Cookie{}
	}
	cookies := make([]*http.Cookie, 0, cookieCount)
	for _, line := range h["Set-Cookie"] {
		parts := strings.Split(strings.TrimSpace(line), ";")
		if len(parts) == 1 && parts[0] == "" {
			continue
		}
		parts[0] = strings.TrimSpace(parts[0])
		j := strings.Index(parts[0], "=")
		if j < 0 {
			continue
		}

		name, value := parts[0][:j], parts[0][j+1:]

		c := &http.Cookie{
			Name:  name,
			Value: value,
			Raw:   line,
		}

		readCookiesParts(c, parts)
		cookies = append(cookies, c)
	}
	return cookies
}

func readCookiesParts(c *http.Cookie, parts []string) {
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.TrimSpace(parts[i])
		if len(parts[i]) == 0 {
			continue
		}
		attr, val := parts[i], ""
		if j := strings.Index(attr, "="); j >= 0 {
			attr, val = attr[:j], attr[j+1:]
		}
		lowerAttr := strings.ToLower(attr)
		switch lowerAttr {
		case "secure":
			c.Secure = true
			continue
		case "httponly":
			c.HttpOnly = true
			continue
		case "domain":
			c.Domain = val
			continue
		case "max-age":
			secs, err := strconv.Atoi(val)
			if err != nil {
				break
			}
			c.MaxAge = secs
			continue
		case "expires":
			c.RawExpires = val
			exptime, err := time.Parse(time.RFC1123, val)
			if err != nil {
				exptime, err = time.Parse("Mon, 02-Jan-2006 15:04:05 MST", val)
				if err != nil {
					c.Expires = time.Time{}
					break
				}
			}
			c.Expires = exptime.UTC()
			continue
		case "path":
			c.Path = val
			continue
		}
		c.Unparsed = append(c.Unparsed, parts[i])
	}
}
