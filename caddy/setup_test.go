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

func TestSetup_CornerCasesJWTSecret(t *testing.T) {

	os.Setenv("JWT_SECRET", "jwtsecret")

	for j, test := range []struct {
		description           string
		envInput              string
		config1               string
		config2               string
		expectedEnv           string
		expectedSecretConfig1 string
		expectedSecretConfig2 string
	}{
		{
			description: "just use the environment",
			envInput:    "foo",
			config1: `login {
                                        simple bob=secret
                                }`,
			config2: `login {
                                        simple bob=secret
                                }`,
			expectedEnv:           "foo",
			expectedSecretConfig1: "foo",
			expectedSecretConfig2: "foo",
		},
		{
			description: "set variable using configs",
			envInput:    "",
			config1: `login {
                                        simple bob=secret
                                        jwt_secret xxx
                                }`,
			config2: `login {
                                        simple bob=secret
                                        jwt_secret yyy
                                }`,
			expectedEnv:           "xxx",
			expectedSecretConfig1: "xxx",
			expectedSecretConfig2: "yyy",
		},
		{
			description: "secret in env and configs was set",
			envInput:    "bli",
			config1: `login {
                                        simple bob=secret
                                        jwt_secret bla
                                }`,
			config2: `login {
                                        simple bob=secret
                                        jwt_secret blub
                                }`,
			expectedEnv:           "bli", // should not be touched
			expectedSecretConfig1: "bla",
			expectedSecretConfig2: "blub",
		},
		{
			description: "random default value",
			envInput:    "",
			config1: `login {
                                        simple bob=secret
                                }`,
			config2: `login {
                                        simple bob=secret
                                }`,
			expectedEnv:           login.DefaultConfig().JwtSecret,
			expectedSecretConfig1: login.DefaultConfig().JwtSecret,
			expectedSecretConfig2: login.DefaultConfig().JwtSecret,
		},
	} {
		t.Run(fmt.Sprintf("test %v %v", j, test.description), func(t *testing.T) {
			if test.envInput == "" {
				os.Unsetenv("JWT_SECRET")
			} else {
				os.Setenv("JWT_SECRET", test.envInput)
			}
			c1 := caddy.NewTestController("http", test.config1)
			NoError(t, setup(c1))
			c2 := caddy.NewTestController("http", test.config2)
			NoError(t, setup(c2))

			mids1 := httpserver.GetConfig(c1).Middleware()
			if len(mids1) == 0 {
				t.Errorf("no middlewares created in test #%v", j)
				return
			}
			middleware1 := mids1[len(mids1)-1](nil).(*CaddyHandler)

			mids2 := httpserver.GetConfig(c2).Middleware()
			if len(mids2) == 0 {
				t.Errorf("no middlewares created in test #%v", j)
				return
			}
			middleware2 := mids2[len(mids2)-1](nil).(*CaddyHandler)

			Equal(t, test.expectedSecretConfig1, middleware1.config.JwtSecret)
			Equal(t, test.expectedSecretConfig2, middleware2.config.JwtSecret)
			Equal(t, test.expectedEnv, os.Getenv("JWT_SECRET"))
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
