package login

import (
	"errors"
	"flag"
	"fmt"
	"github.com/tarent/lib-compose/logging"
	"github.com/tarent/loginsrv/oauth2"
	"math/rand"
	"os"
	"strings"
	"time"
)

var jwtDefaultSecret string

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	jwtDefaultSecret = randStringBytes(32)
}

func DefaultConfig() *Config {
	return &Config{
		Host:           "localhost",
		Port:           "6789",
		LogLevel:       "info",
		JwtSecret:      jwtDefaultSecret,
		SuccessUrl:     "/",
		LogoutUrl:      "",
		LoginPath:      "/login",
		CookieName:     "jwt_token",
		CookieHttpOnly: true,
		Backends:       Options{},
		Oauth:          Options{},
	}
}

const envPrefix = "LOGINSRV_"

type Config struct {
	Host           string
	Port           string
	LogLevel       string
	TextLogging    bool
	JwtSecret      string
	SuccessUrl     string
	LogoutUrl      string
	Template       string
	LoginPath      string
	CookieName     string
	CookieHttpOnly bool
	Backends       Options
	Oauth          Options
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

// ConfigureFlagSet adds all flags to the supplied flag set
func (c *Config) ConfigureFlagSet(f *flag.FlagSet) {
	f.StringVar(&c.Host, "host", c.Host, "The host to listen on")
	f.StringVar(&c.Port, "port", c.Port, "The port to listen on")
	f.StringVar(&c.LogLevel, "log-level", c.LogLevel, "The log level")
	f.BoolVar(&c.TextLogging, "text-logging", c.TextLogging, "Log in text format instead of json")
	f.StringVar(&c.JwtSecret, "jwt-secret", "random key", "The secret to sign the jwt token")
	f.StringVar(&c.CookieName, "cookie-name", c.CookieName, "The name of the jwt cookie")
	f.BoolVar(&c.CookieHttpOnly, "cookie-http-only", c.CookieHttpOnly, "Set the cookie with the http only flag")
	f.StringVar(&c.SuccessUrl, "success-url", c.SuccessUrl, "The url to redirect after login")
	f.StringVar(&c.LogoutUrl, "logout-url", c.LogoutUrl, "The url or path to redirect after logout")
	f.StringVar(&c.Template, "template", c.Template, "An alternative template for the login form")
	f.StringVar(&c.LoginPath, "login-path", c.LoginPath, "The path of the login resource")

	// the -backends is deprecated, but we support it for backwards compatibility
	deprecatedBackends := setFunc(func(optsKvList string) error {
		logging.Logger.Warn("DEPRECATED: '-backend' is no loger supported. Please set the backends by explicit paramters")
		opts, err := parseOptions(optsKvList)
		if err != nil {
			return err
		}
		pName, ok := opts["provider"]
		if !ok {
			return errors.New("missing provder name provider=..")
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

	// prefer environment settings
	f.VisitAll(func(f *flag.Flag) {
		if val, isPresent := os.LookupEnv(envName(f.Name)); isPresent {
			f.Value.Set(val)
		}
	})

	err := f.Parse(args)
	if err != nil {
		return nil, err
	}

	if config.JwtSecret == "random key" {
		if s, set := os.LookupEnv("LOGINSRV_JWT_SECRET"); set {
			config.JwtSecret = s
		} else {
			config.JwtSecret = jwtDefaultSecret
		}
	}

	return config, err
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
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
