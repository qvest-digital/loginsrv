package caddy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/login"
)

func TestSetup(t *testing.T) {

	os.Setenv("JWT_SECRET", "jwtsecret")

	for j, test := range []struct {
		input       string
		shouldErr   bool
		config      login.Config
		configCheck func(*testing.T, *login.Config)
	}{
		{
			input: `login {
                                        simple bob=secret
                                }`,
			shouldErr: false,
			configCheck: func(t *testing.T, cfg *login.Config) {
				expectedBackendCfg := login.Options{"simple": map[string]string{"bob": "secret"}}
				Equal(t, expectedBackendCfg, cfg.Backends, "config simple auth backend")
			},
		},
		{
			input: `login {
                                        success_url successurl
                                        jwt_expiry 42h
                                        jwt_algo algo
                                        login_path /foo/bar
                                        redirect true
                                        redirect_query_parameter comingFrom
                                        redirect_check_referer true
                                        redirect_host_file domainWhitelist.txt
                                        cookie_name cookiename
                                        cookie_http_only false
                                        cookie_domain example.com
                                        cookie_expiry 23h23m
                                        simple bob=secret
                                        osiam endpoint=http://localhost:8080,client_id=example-client,client_secret=secret
                                }`,
			shouldErr: false,
			configCheck: func(t *testing.T, cfg *login.Config) {
				Equal(t, cfg.SuccessURL, "successurl")
				Equal(t, cfg.JwtExpiry, 42*time.Hour)
				Equal(t, cfg.JwtAlgo, "algo")
				Equal(t, cfg.LoginPath, "/foo/bar")
				Equal(t, cfg.Redirect, true)
				Equal(t, cfg.RedirectQueryParameter, "comingFrom")
				Equal(t, cfg.RedirectCheckReferer, true)
				Equal(t, cfg.RedirectHostFile, "domainWhitelist.txt")
				Equal(t, cfg.CookieName, "cookiename")
				Equal(t, cfg.CookieHTTPOnly, false)
				Equal(t, cfg.CookieDomain, "example.com")
				Equal(t, cfg.CookieExpiry, 23*time.Hour+23*time.Minute)
				expectedBackendCfg := login.Options{
					"simple": map[string]string{
						"bob": "secret",
					},
					"osiam": map[string]string{
						"endpoint":      "http://localhost:8080",
						"client_id":     "example-client",
						"client_secret": "secret",
					},
				}
				Equal(t, expectedBackendCfg, cfg.Backends, "config simple auth backend")
			},
		},
		{
			input: `loginsrv /context {
                                        backend provider=simple,bob=secret
                                        cookie-name cookiename
                                }`,
			shouldErr: false,
			configCheck: func(t *testing.T, cfg *login.Config) {
				Equal(t, "/context/login", cfg.LoginPath, "Login path should be set by argument for backwards compatibility")
				Equal(t, "cookiename", cfg.CookieName, "The cookie name should be set by a config name with - instead of _ for backwards compatibility")
				expectedBackendCfg := login.Options{
					"simple": map[string]string{
						"bob": "secret",
					},
				}
				Equal(t, expectedBackendCfg, cfg.Backends, "The backend config should be set by \"backend provider=\" for backwards compatibility")
			},
		},
		// error cases
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
				Error(t, err, "test ")
				return
			}
			NoError(t, err)
			mids := httpserver.GetConfig(c).Middleware()
			if len(mids) == 0 {
				t.Errorf("no middlewares created in test #%v", j)
				return
			}
			middleware := mids[len(mids)-1](nil).(*CaddyHandler)
			test.configCheck(t, middleware.config)
		})
	}
}

func TestSetup_RelativeFiles(t *testing.T) {
	caddyfile := `loginsrv {
                        template myTemplate.tpl
                        redirect_host_file redirectDomains.txt
                        simple bob=secret
                      }`
	root, _ := ioutil.TempDir("", "")

	c := caddy.NewTestController("http", caddyfile)
	c.Key = "RelativeTemplateFileTest"
	config := httpserver.GetConfig(c)
	config.Root = root

	err := setup(c)
	NoError(t, err)
	mids := httpserver.GetConfig(c).Middleware()
	if len(mids) == 0 {
		t.Errorf("no middlewares created")
	}
	middleware := mids[len(mids)-1](nil).(*CaddyHandler)

	Equal(t, filepath.FromSlash(root+"/myTemplate.tpl"), middleware.config.Template)
	Equal(t, "redirectDomains.txt", middleware.config.RedirectHostFile)
}
