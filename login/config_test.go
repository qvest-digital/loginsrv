package login

import (
	"flag"
	"os"
	"testing"
	"time"

	. "github.com/stretchr/testify/assert"
)

func TestConfig_ReadConfigDefaults(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	defaultConfig := DefaultConfig()
	gotConfig := ReadConfig()
	defaultConfig.JwtSecret = "random"
	gotConfig.JwtSecret = "random"
	Equal(t, defaultConfig, gotConfig)
}

func TestConfig_ReadConfig(t *testing.T) {
	input := []string{
		"--host=host",
		"--port=port",
		"--log-level=loglevel",
		"--text-logging=true",
		"--jwt-secret=jwtsecret",
		"--jwt-algo=algo",
		"--jwt-expiry=42h42m",
		"--success-url=successurl",
		"--redirect=false",
		"--redirect-query-parameter=comingFrom",
		"--redirect-check-referer=false",
		"--redirect-host-file=File",
		"--logout-url=logouturl",
		"--template=template",
		"--login-path=loginpath",
		"--cookie-name=cookiename",
		"--cookie-expiry=23m",
		"--cookie-domain=*.example.com",
		"--cookie-http-only=false",
		"--backend=provider=simple",
		"--backend=provider=foo",
		"--github=client_id=foo,client_secret=bar",
		"--grace-period=4s",
		"--user-file=users.yml",
	}

	expected := &Config{
		Host:                   "host",
		Port:                   "port",
		LogLevel:               "loglevel",
		TextLogging:            true,
		JwtSecret:              "jwtsecret",
		JwtAlgo:                "algo",
		JwtExpiry:              42*time.Hour + 42*time.Minute,
		SuccessURL:             "successurl",
		Redirect:               false,
		RedirectQueryParameter: "comingFrom",
		RedirectCheckReferer:   false,
		RedirectHostFile:       "File",
		LogoutURL:              "logouturl",
		Template:               "template",
		LoginPath:              "loginpath",
		CookieName:             "cookiename",
		CookieExpiry:           23 * time.Minute,
		CookieDomain:           "*.example.com",
		CookieHTTPOnly:         false,
		Backends: Options{
			"simple": map[string]string{},
			"foo":    map[string]string{},
		},
		Oauth: Options{
			"github": map[string]string{
				"client_id":     "foo",
				"client_secret": "bar",
			},
		},
		GracePeriod: 4 * time.Second,
		UserFile:    "users.yml",
	}

	cfg, err := readConfig(flag.NewFlagSet("", flag.ContinueOnError), input)
	NoError(t, err)
	Equal(t, expected, cfg)
}

func TestConfig_ReadConfigFromEnv(t *testing.T) {
	NoError(t, os.Setenv("LOGINSRV_HOST", "host"))
	NoError(t, os.Setenv("LOGINSRV_PORT", "port"))
	NoError(t, os.Setenv("LOGINSRV_LOG_LEVEL", "loglevel"))
	NoError(t, os.Setenv("LOGINSRV_TEXT_LOGGING", "true"))
	NoError(t, os.Setenv("LOGINSRV_JWT_SECRET", "jwtsecret"))
	NoError(t, os.Setenv("LOGINSRV_JWT_ALGO", "algo"))
	NoError(t, os.Setenv("LOGINSRV_JWT_EXPIRY", "42h42m"))
	NoError(t, os.Setenv("LOGINSRV_SUCCESS_URL", "successurl"))
	NoError(t, os.Setenv("LOGINSRV_REDIRECT", "false"))
	NoError(t, os.Setenv("LOGINSRV_REDIRECT_QUERY_PARAMETER", "comingFrom"))
	NoError(t, os.Setenv("LOGINSRV_REDIRECT_CHECK_REFERER", "false"))
	NoError(t, os.Setenv("LOGINSRV_REDIRECT_HOST_FILE", "File"))
	NoError(t, os.Setenv("LOGINSRV_LOGOUT_URL", "logouturl"))
	NoError(t, os.Setenv("LOGINSRV_TEMPLATE", "template"))
	NoError(t, os.Setenv("LOGINSRV_LOGIN_PATH", "loginpath"))
	NoError(t, os.Setenv("LOGINSRV_COOKIE_NAME", "cookiename"))
	NoError(t, os.Setenv("LOGINSRV_COOKIE_EXPIRY", "23m"))
	NoError(t, os.Setenv("LOGINSRV_COOKIE_DOMAIN", "*.example.com"))
	NoError(t, os.Setenv("LOGINSRV_COOKIE_HTTP_ONLY", "false"))
	NoError(t, os.Setenv("LOGINSRV_SIMPLE", "foo=bar"))
	NoError(t, os.Setenv("LOGINSRV_GITHUB", "client_id=foo,client_secret=bar"))
	NoError(t, os.Setenv("LOGINSRV_GRACE_PERIOD", "4s"))

	expected := &Config{
		Host:                   "host",
		Port:                   "port",
		LogLevel:               "loglevel",
		TextLogging:            true,
		JwtSecret:              "jwtsecret",
		JwtAlgo:                "algo",
		JwtExpiry:              42*time.Hour + 42*time.Minute,
		SuccessURL:             "successurl",
		Redirect:               false,
		RedirectQueryParameter: "comingFrom",
		RedirectCheckReferer:   false,
		RedirectHostFile:       "File",
		LogoutURL:              "logouturl",
		Template:               "template",
		LoginPath:              "loginpath",
		CookieName:             "cookiename",
		CookieExpiry:           23 * time.Minute,
		CookieDomain:           "*.example.com",
		CookieHTTPOnly:         false,
		Backends: Options{
			"simple": map[string]string{
				"foo": "bar",
			},
		},
		Oauth: Options{
			"github": map[string]string{
				"client_id":     "foo",
				"client_secret": "bar",
			},
		},
		GracePeriod: 4 * time.Second,
	}

	cfg, err := readConfig(flag.NewFlagSet("", flag.ContinueOnError), []string{})
	NoError(t, err)
	Equal(t, expected, cfg)
}
