package caddy

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/tarent/loginsrv/login"
	"github.com/tarent/loginsrv/model"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

//Tests a page while being logged in as a user (doesn't test that the {user} replacer changes)
func Test_ServeHTTP_200(t *testing.T) {
	//Set the ServeHTTP *http.Request
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Unable to create request: %v", err)
	}

	/**
	TODO: This will only work with the caddy master branch or the next caddy release

	// Associate a replacer with the request:
	r = r.WithContext(context.WithValue(context.Background(), httpserver.ReplacerCtxKey, httpserver.NewReplacer(r, nil, "-")))
	*/

	//Set the ServeHTTP http.ResponseWriter
	w := httptest.NewRecorder()

	//Set the CaddyHandler config
	configh := login.DefaultConfig()
	configh.Backends = login.Options{"simple": {"bob": "secret"}}
	loginh, err := login.NewHandler(configh)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	//Set the CaddyHandler that will use ServeHTTP
	h := &CaddyHandler{
		next: httpserver.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			return http.StatusOK, nil // not t.Fatalf, or we will not see what other methods yield
		}),
		config:       login.DefaultConfig(),
		loginHandler: loginh,
	}

	//Set user token
	userInfo := model.UserInfo{Sub: "bob", Expiry: time.Now().Add(time.Second).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, userInfo)
	validToken, err := token.SignedString([]byte(h.config.JwtSecret))
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	//Set cookie for user token on the ServeHTTP http.ResponseWriter
	cookie := http.Cookie{Name: "jwt_token", Value: validToken, HttpOnly: true}
	http.SetCookie(w, &cookie)

	//Add the cookie to the request
	r.AddCookie(&cookie)

	//Test that cookie is a valid token
	_, valid := loginh.GetToken(r)
	if !valid {
		t.Errorf("loginHandler cookie is not valid")
	}

	status, err := h.ServeHTTP(w, r)

	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if status != 200 {
		t.Errorf("Expected returned status code to be %d, got %d", 0, status)
	}

	/**
	TODO: This will only work with the caddy master branch or the next caddy release


		// Check that the replacer now is able to substitute the user variable in log lines
		replacer, replacerOk := r.Context().Value(httpserver.ReplacerCtxKey).(httpserver.Replacer)
		if !replacerOk {
			t.Errorf("no replacer associated with request")

		} else {
			replacement := replacer.Replace("{user}")
			if replacement != "bob" {
				t.Errorf(`wrong replacement: expected "bob", but got %q`, replacement)
			}
		}
	*/
}

//Tests the login page without being logged as a user (doesn't test that the {user} replacer stays as-is)
func Test_ServeHTTP_login(t *testing.T) {
	//Set the ServeHTTP *http.Request
	r, err := http.NewRequest("GET", "/login", nil)
	if err != nil {
		t.Fatalf("Unable to create request: %v", err)
	}

	//Set the ServeHTTP http.ResponseWriter
	w := httptest.NewRecorder()

	//Set the CaddyHandler config
	configh := login.DefaultConfig()
	configh.Backends = login.Options{"simple": {"bob": "secret"}}
	loginh, err := login.NewHandler(configh)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	//Set the CaddyHandler that will use ServeHTTP
	h := &CaddyHandler{
		next: httpserver.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			return http.StatusOK, nil // not t.Fatalf, or we will not see what other methods yield
		}),
		config:       login.DefaultConfig(),
		loginHandler: loginh,
	}

	status, err := h.ServeHTTP(w, r)

	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if status != 0 {
		t.Errorf("Expected returned status code to be %d, got %d", 0, status)
	}
}
