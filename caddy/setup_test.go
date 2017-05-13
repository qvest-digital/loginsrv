package caddy

import (
	"fmt"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/login"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestSetup(t *testing.T) {

	os.Setenv("JWT_SECRET", "jwtsecret")

	for j, test := range []struct {
		input     string
		shouldErr bool
		config    login.Config
	}{
		{ //defaults
			input: `login {
                                        simple bob=secret
                                }`,
			shouldErr: false,
			config: login.Config{
				JwtSecret:      "jwtsecret",
				JwtExpiry:      24 * time.Hour,
				SuccessUrl:     "/",
				LoginPath:      "/login",
				CookieName:     "jwt_token",
				CookieHttpOnly: true,
				Backends: login.Options{
					"simple": map[string]string{
						"bob": "secret",
					},
				},
				Oauth: login.Options{},
			}},
		{
			input: `login {
                                        success_url successurl
                                        jwt_expiry 42h
                                        login_path /foo/bar
                                        cookie_name cookiename
                                        cookie_http_only false
                                        cookie_domain example.com
                                        cookie_expiry 23h23m
                                        simple bob=secret
                                        osiam endpoint=http://localhost:8080,client_id=example-client,client_secret=secret
                                }`,
			shouldErr: false,
			config: login.Config{
				JwtSecret:      "jwtsecret",
				JwtExpiry:      42 * time.Hour,
				SuccessUrl:     "successurl",
				LoginPath:      "/foo/bar",
				CookieName:     "cookiename",
				CookieDomain:   "example.com",
				CookieExpiry:   23*time.Hour + 23*time.Minute,
				CookieHttpOnly: false,
				Backends: login.Options{
					"simple": map[string]string{
						"bob": "secret",
					},
					"osiam": map[string]string{
						"endpoint":      "http://localhost:8080",
						"client_id":     "example-client",
						"client_secret": "secret",
					},
				},
				Oauth: login.Options{},
			}},
		{ // backwards compatibility
			// * login path as argument
			// * '-' in parameter names
			// * backend config by 'backend provider='
			input: `loginsrv /context {
                                        backend provider=simple,bob=secret
                                        cookie-name cookiename
                                }`,
			shouldErr: false,
			config: login.Config{
				JwtSecret:      "jwtsecret",
				JwtExpiry:      24 * time.Hour,
				SuccessUrl:     "/",
				LoginPath:      "/context/login",
				CookieName:     "cookiename",
				CookieHttpOnly: true,
				Backends: login.Options{
					"simple": map[string]string{
						"bob": "secret",
					},
				},
				Oauth: login.Options{},
			}},
		{ // backwards compatibility
			// * login path as argument
			// * '-' in parameter names
			// * backend config by 'backend provider='
			input: `loginsrv / {
                                        backend provider=simple,bob=secret
                                        cookie-name cookiename
                                }`,
			shouldErr: false,
			config: login.Config{
				JwtSecret:      "jwtsecret",
				JwtExpiry:      24 * time.Hour,
				SuccessUrl:     "/",
				LoginPath:      "/login",
				CookieName:     "cookiename",
				CookieHttpOnly: true,
				Backends: login.Options{
					"simple": map[string]string{
						"bob": "secret",
					},
				},
				Oauth: login.Options{},
			}},

		// error cases
		{ // duration parse error
			input: `login {
                                        simple bob=secret
                                }`,
			shouldErr: false,
			config: login.Config{
				JwtSecret:      "jwtsecret",
				JwtExpiry:      24 * time.Hour,
				SuccessUrl:     "/",
				LoginPath:      "/login",
				CookieName:     "jwt_token",
				CookieHttpOnly: true,
				Backends: login.Options{
					"simple": map[string]string{
						"bob": "secret",
					},
				},
				Oauth: login.Options{},
			}},
		{input: "login {\n}", shouldErr: true},
		{input: "login xx yy {\n}", shouldErr: true},
		{input: "login {\n cookie_http_only 42d \n simple bob=secret \n}", shouldErr: true},
		{input: "login {\n unknown property \n simple bob=secret \n}", shouldErr: true},
		{input: "login {\n backend \n}", shouldErr: true},
		{input: "login {\n backend provider=foo\n}", shouldErr: true},
		{input: "login {\n backend kk\n}", shouldErr: true},
	} {
		t.Run(fmt.Sprintf("test %v", j), func(t *testing.T) {
			c := caddy.NewTestController("http", test.input)
			err := setup(c)
			if test.shouldErr {
				assert.Error(t, err, "test ")
				return
			} else {
				assert.NoError(t, err)
			}
			mids := httpserver.GetConfig(c).Middleware()
			if len(mids) == 0 {
				t.Errorf("no middlewares created in test #%v", j)
				return
			}
			middleware := mids[len(mids)-1](nil).(*CaddyHandler)
			assert.Equal(t, &test.config, middleware.config)
		})
	}
}

func TestSetup_RelativeTemplateFile(t *testing.T) {
	caddyfile := "loginsrv {\n  template myTemplate.tpl\n  simple bob=secret\n}"
	root, _ := ioutil.TempDir("", "")
	expectedPath := root + "/myTemplate.tpl"

	c := caddy.NewTestController("http", caddyfile)
	c.Key = "RelativeTemplateFileTest"
	config := httpserver.GetConfig(c)
	config.Root = root

	err := setup(c)
	assert.NoError(t, err)
	mids := httpserver.GetConfig(c).Middleware()
	if len(mids) == 0 {
		t.Errorf("no middlewares created")
	}
	middleware := mids[len(mids)-1](nil).(*CaddyHandler)

	assert.Equal(t, expectedPath, middleware.config.Template)
}
