package login

import (
	"flag"
	. "github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
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
		"--jwt-expiry=42h42m",
		"--success-url=successurl",
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
	}

	expected := &Config{
		Host:           "host",
		Port:           "port",
		LogLevel:       "loglevel",
		TextLogging:    true,
		JwtSecret:      "jwtsecret",
		JwtExpiry:      42*time.Hour + 42*time.Minute,
		SuccessURL:     "successurl",
		LogoutURL:      "logouturl",
		Template:       "template",
		LoginPath:      "loginpath",
		CookieName:     "cookiename",
		CookieExpiry:   23 * time.Minute,
		CookieDomain:   "*.example.com",
		CookieHTTPOnly: false,
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
	NoError(t, os.Setenv("LOGINSRV_JWT_EXPIRY", "42h42m"))
	NoError(t, os.Setenv("LOGINSRV_SUCCESS_URL", "successurl"))
	NoError(t, os.Setenv("LOGINSRV_LOGOUT_URL", "logouturl"))
	NoError(t, os.Setenv("LOGINSRV_TEMPLATE", "template"))
	NoError(t, os.Setenv("LOGINSRV_LOGIN_PATH", "loginpath"))
	NoError(t, os.Setenv("LOGINSRV_COOKIE_NAME", "cookiename"))
	NoError(t, os.Setenv("LOGINSRV_COOKIE_EXPIRY", "23m"))
	NoError(t, os.Setenv("LOGINSRV_COOKIE_DOMAIN", "*.example.com"))
	NoError(t, os.Setenv("LOGINSRV_COOKIE_HTTP_ONLY", "false"))
	NoError(t, os.Setenv("LOGINSRV_SIMPLE", "foo=bar"))
	NoError(t, os.Setenv("LOGINSRV_GITHUB", "client_id=foo,client_secret=bar"))

	expected := &Config{
		Host:           "host",
		Port:           "port",
		LogLevel:       "loglevel",
		TextLogging:    true,
		JwtSecret:      "jwtsecret",
		JwtExpiry:      42*time.Hour + 42*time.Minute,
		SuccessURL:     "successurl",
		LogoutURL:      "logouturl",
		Template:       "template",
		LoginPath:      "loginpath",
		CookieName:     "cookiename",
		CookieExpiry:   23 * time.Minute,
		CookieDomain:   "*.example.com",
		CookieHTTPOnly: false,
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
	}

	cfg, err := readConfig(flag.NewFlagSet("", flag.ContinueOnError), []string{})
	NoError(t, err)
	Equal(t, expected, cfg)
}
