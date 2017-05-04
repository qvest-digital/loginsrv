package caddy

import (
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/login"
	"os"
	"testing"
)

func TestSetup(t *testing.T) {

	os.Setenv("JWT_SECRET", "jwtsecret")

	for j, test := range []struct {
		input     string
		shouldErr bool
		config    login.Config
	}{
		{ //defaults
			input: `loginsrv {
                                        simple bob=secret
                                }`,
			shouldErr: false,
			config: login.Config{
				JwtSecret:      "jwtsecret",
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
			input: `loginsrv {
                                        success_url successurl
                                        login_path /foo/bar
                                        cookie_name cookiename
                                        cookie_http_only false
                                        simple bob=secret
                                        osiam endpoint=http://localhost:8080,client_id=example-client,client_secret=secret
                                }`,
			shouldErr: false,
			config: login.Config{
				JwtSecret:      "jwtsecret",
				SuccessUrl:     "successurl",
				LoginPath:      "/foo/bar",
				CookieName:     "cookiename",
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
		{input: "loginsrv {\n}", shouldErr: true},
		{input: "loginsrv xx yy {\n}", shouldErr: true},
		{input: "loginsrv {\n cookie_http_only 42 \n simple bob=secret \n}", shouldErr: true},
		{input: "loginsrv {\n unknown property \n simple bob=secret \n}", shouldErr: true},
		{input: "loginsrv {\n backend \n}", shouldErr: true},
		{input: "loginsrv {\n backend provider=foo\n}", shouldErr: true},
		{input: "loginsrv {\n backend kk\n}", shouldErr: true},
	} {
		c := caddy.NewTestController("http", test.input)
		err := setup(c)
		if err != nil && !test.shouldErr {
			t.Errorf("Test case #%d received an error of %v", j, err)
		} else if test.shouldErr {
			continue
		}
		mids := httpserver.GetConfig(c).Middleware()
		if len(mids) == 0 {
			t.Errorf("no middlewares created in test #%v", j)
			continue
		}
		middleware := mids[len(mids)-1](nil).(*CaddyHandler)
		assert.Equal(t, &test.config, middleware.config)
	}
}
