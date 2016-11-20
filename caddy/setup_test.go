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
		path      string
		config    login.Config
	}{
		{ //defaults
			input: `loginsrv / {
                                        backend provider=simple,bob=secret
                                }`,
			shouldErr: false,
			path:      "/",
			config: login.Config{
				JwtSecret:      "jwtsecret",
				SuccessUrl:     "/",
				CookieName:     "jwt_token",
				CookieHttpOnly: true,
				Backends: login.BackendOptions{
					map[string]string{
						"provider": "simple",
						"bob":      "secret",
					},
				},
			}},
		{
			input: `loginsrv / {
                                        success-url successurl
                                        cookie-name cookiename
                                        cookie-http-only false
                                        backend provider=simple,bob=secret
                                        backend provider=osiam,endpoint=http://localhost:8080,clientId=example-client,clientSecret=secret
                                }`,
			shouldErr: false,
			path:      "/",
			config: login.Config{
				JwtSecret:      "jwtsecret",
				SuccessUrl:     "successurl",
				CookieName:     "cookiename",
				CookieHttpOnly: false,
				Backends: login.BackendOptions{
					map[string]string{
						"provider": "simple",
						"bob":      "secret",
					},
					map[string]string{
						"provider":     "osiam",
						"endpoint":     "http://localhost:8080",
						"clientId":     "example-client",
						"clientSecret": "secret",
					},
				},
			}},
		// error cases
		{input: "loginsrv {\n}", shouldErr: true},
		{input: "loginsrv xx yy {\n}", shouldErr: true},
		{input: "loginsrv / {\n cookie-http-only 42 \n backend provider=simple,bob=secret \n}", shouldErr: true},
		{input: "loginsrv / {\n unknown property \n backend provider=simple,bob=secret \n}", shouldErr: true},
		{input: "loginsrv / {\n backend \n}", shouldErr: true},
		{input: "loginsrv / {\n backend provider=foo\n}", shouldErr: true},
		{input: "loginsrv / {\n backend kk\n}", shouldErr: true},
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
		assert.Equal(t, test.path, middleware.path)
		assert.Equal(t, &test.config, middleware.config)
	}
}
