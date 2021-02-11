package login

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/tarent/loginsrv/logging"
	"github.com/tarent/loginsrv/oauth2"
)

var jwtDefaultSecret string

func init() {
	var err error
	jwtDefaultSecret, err = randStringBytes(64)
	if err != nil {
		panic(err)
	}
}

// DefaultConfig for the loginsrv handler
func DefaultConfig() *Config {
	return &Config{
		Host:                   "localhost",
		Port:                   "6789",
		LogLevel:               "info",
		JwtSecret:              jwtDefaultSecret,
		JwtAlgo:                "HS512",
		JwtExpiry:              24 * time.Hour,
		JwtRefreshes:           0,
		JwtKeyID:               "",
		SuccessURL:             "/",
		Redirect:               true,
		RedirectQueryParameter: "backTo",
		RedirectCheckReferer:   true,
		RedirectHostFile:       "",
		LogoutURL:              "",
		LoginPath:              "/login",
		CookieName:             "jwt_token",
		CookieHTTPOnly:         true,
		CookieSecure:           true,
		Backends:               Options{},
		Oauth:                  Options{},
		GracePeriod:            5 * time.Second,
		UserFile:               "",
		UserEndpoint:           "",
		UserEndpointToken:      "",
		UserEndpointTimeout:    5 * time.Second,
	}
}

const envPrefix = "LOGINSRV_"

// Config for the loginsrv handler
type Config struct {
	Host                   string
	Port                   string
	LogLevel               string
	TextLogging            bool
	JwtSecret              string
	JwtSecretFile          string
	JwtAlgo                string
	JwtExpiry              time.Duration
	JwtRefreshes           int
	JwtKeyID               string
	SuccessURL             string
	Redirect               bool
	RedirectQueryParameter string
	RedirectCheckReferer   bool
	RedirectHostFile       string
	LogoutURL              string
	Template               string
	LoginPath              string
	CookieName             string
	CookieExpiry           time.Duration
	CookieDomain           string
	CookieHTTPOnly         bool
	CookieSecure           bool
	Backends               Options
	Oauth                  Options
	GracePeriod            time.Duration
	UserFile               string
	UserEndpoint           string
	UserEndpointToken      string
	UserEndpointTimeout    time.Duration
}

// Options is the configuration structure for oauth and backend provider
// key is the providername, value is a options map.
type Options map[string]map[string]string

// addOauthOpts adds the options for a provider in the form of key=value,key=value,..
func (c *Config) addOauthOpts(providerName, optsKvList string) error {
	opts, err := parseOptions(optsKvList)
	if err != nil {
		return err
	}

	c.Oauth[providerName] = opts
	return nil
}

// addBackendOpts adds the options for a provider in the form of key=value,key=value,..
func (c *Config) addBackendOpts(providerName, optsKvList string) error {
	opts, err := parseOptions(optsKvList)
	if err != nil {
		return err
	}

	c.Backends[providerName] = opts
	return nil
}

// ResolveFileReferences resolves configuration values, which are dynamically referenced via files
func (c *Config) ResolveFileReferences() error {
	// Try to load the secret from a file, if set
	if c.JwtSecretFile != "" {
		secretBytes, err := ioutil.ReadFile(c.JwtSecretFile)
		if err != nil {
			return err
		}

		c.JwtSecret = string(secretBytes)
	}

	return nil
}

// ConfigureFlagSet adds all flags to the supplied flag set
func (c *Config) ConfigureFlagSet(f *flag.FlagSet) {
	f.StringVar(&c.Host, "host", c.Host, "The host to listen on")
	f.StringVar(&c.Port, "port", c.Port, "The port to listen on")
	f.StringVar(&c.LogLevel, "log-level", c.LogLevel, "The log level")
	f.BoolVar(&c.TextLogging, "text-logging", c.TextLogging, "Log in text format instead of json")
	f.StringVar(&c.JwtSecret, "jwt-secret", c.JwtSecret, "The secret to sign the jwt token")
	f.StringVar(&c.JwtSecretFile, "jwt-secret-file", c.JwtSecretFile, "Path to a file containing the secret to sign the jwt token (overrides jwt-secret)")
	f.StringVar(&c.JwtAlgo, "jwt-algo", c.JwtAlgo, "The signing algorithm to use (ES256, ES384, ES512, RS256, RS384, RS512, HS256, HS384, HS512)")
	f.DurationVar(&c.JwtExpiry, "jwt-expiry", c.JwtExpiry, "The expiry duration for the jwt token, e.g. 2h or 3h30m")
	f.IntVar(&c.JwtRefreshes, "jwt-refreshes", c.JwtRefreshes, "The maximum amount of jwt refreshes. 0 by Default")
	f.StringVar(&c.JwtKeyID, "jwt-key-id", c.JwtKeyID, "The key id to use, added to the jwt header as kid if set")
	f.StringVar(&c.CookieName, "cookie-name", c.CookieName, "The name of the jwt cookie")
	f.BoolVar(&c.CookieHTTPOnly, "cookie-http-only", c.CookieHTTPOnly, "Set the cookie with the http only flag")
	f.BoolVar(&c.CookieSecure, "cookie-secure", c.CookieSecure, "Set the cookie with the secure flag")
	f.DurationVar(&c.CookieExpiry, "cookie-expiry", c.CookieExpiry, "The expiry duration for the cookie, e.g. 2h or 3h30m. Default is browser session")
	f.StringVar(&c.CookieDomain, "cookie-domain", c.CookieDomain, "The optional domain parameter for the cookie")
	f.StringVar(&c.SuccessURL, "success-url", c.SuccessURL, "The url to redirect after login")
	f.BoolVar(&c.Redirect, "redirect", c.Redirect, "Allow dynamic overwriting of the the success by query parameter")
	f.StringVar(&c.RedirectQueryParameter, "redirect-query-parameter", c.RedirectQueryParameter, "URL parameter for the redirect target")
	f.BoolVar(&c.RedirectCheckReferer, "redirect-check-referer", c.RedirectCheckReferer, "When redirecting check that the referer is the same domain")
	f.StringVar(&c.RedirectHostFile, "redirect-host-file", c.RedirectHostFile, "A file containing a list of domains that redirects are allowed to, one domain per line")

	f.StringVar(&c.LogoutURL, "logout-url", c.LogoutURL, "The url or path to redirect after logout")
	f.StringVar(&c.Template, "template", c.Template, "An alternative template for the login form")
	f.StringVar(&c.LoginPath, "login-path", c.LoginPath, "The path of the login resource")
	f.DurationVar(&c.GracePeriod, "grace-period", c.GracePeriod, "Graceful shutdown grace period")
	f.StringVar(&c.UserFile, "user-file", c.UserFile, "A YAML file with user specific data for the tokens")
	f.StringVar(&c.UserEndpoint, "user-endpoint", c.UserEndpoint, "URL of an endpoint providing user specific data for the tokens")
	f.StringVar(&c.UserEndpointToken, "user-endpoint-token", c.UserEndpointToken, "Authentication token used when communicating with the user endpoint")
	f.DurationVar(&c.UserEndpointTimeout, "user-endpoint-timeout", c.UserEndpointTimeout, "Timeout used when communicating with the user endpoint")

	// the -backends is deprecated, but we support it for backwards compatibility
	deprecatedBackends := setFunc(func(optsKvList string) error {
		logging.Logger.Warn("DEPRECATED: '-backend' is no longer supported. Please set the backends by explicit parameters")
		opts, err := parseOptions(optsKvList)
		if err != nil {
			return err
		}
		pName, ok := opts["provider"]
		if !ok {
			return errors.New("missing provider name provider=...")
		}
		delete(opts, "provider")
		c.Backends[pName] = opts
		return nil
	})
	f.Var(deprecatedBackends, "backend", "Deprecated, please use the explicit flags")

	// One option for each oauth provider
	for _, pName := range oauth2.ProviderList() {
		func(pName string) {
			setter := setFunc(func(optsKvList string) error {
				return c.addOauthOpts(pName, optsKvList)
			})
			f.Var(setter, pName, "Oauth config in the form: client_id=..,client_secret=..[,scope=..,][redirect_uri=..]")
		}(pName)
	}

	// One option for each backend provider
	for _, pName := range ProviderList() {
		func(pName string) {
			setter := setFunc(func(optsKvList string) error {
				return c.addBackendOpts(pName, optsKvList)
			})
			desc, _ := GetProviderDescription(pName)
			f.Var(setter, pName, desc.HelpText)
		}(pName)
	}
}

// ReadConfig from the commandline args
func ReadConfig() *Config {
	c, err := readConfig(flag.CommandLine, os.Args[1:])
	if err != nil {
		// should never happen, because of flag default policy ExitOnError
		panic(err)
	}
	return c
}

func readConfig(f *flag.FlagSet, args []string) (*Config, error) {
	config := DefaultConfig()
	config.ConfigureFlagSet(f)

	// fist use the environment settings
	f.VisitAll(func(f *flag.Flag) {
		if val, isPresent := os.LookupEnv(envName(f.Name)); isPresent {
			f.Value.Set(val)
		}
	})

	// prefer flags over environment settings
	err := f.Parse(args)
	if err != nil {
		return nil, err
	}

	if err := config.ResolveFileReferences(); err != nil {
		return nil, err
	}

	return config, err
}

func randStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func envName(flagName string) string {
	return envPrefix + strings.Replace(strings.ToUpper(flagName), "-", "_", -1)
}

func parseOptions(b string) (map[string]string, error) {
	opts := map[string]string{}
	pairs := strings.Split(b, ",")
	for _, p := range pairs {
		pair := strings.SplitN(p, "=", 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("provider configuration has to be in form 'key1=value1,key2=..', but was %v", p)
		}
		opts[pair[0]] = pair[1]
	}
	return opts, nil
}

// Helper type to wrap a function closure with the Value interface
type setFunc func(optsKvList string) error

func (f setFunc) Set(value string) error {
	return f(value)
}

func (f setFunc) String() string {
	return "setFunc"
}
