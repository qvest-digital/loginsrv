package login

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env"
	"math/rand"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type BackendOptions []map[string]string

func (bo *BackendOptions) String() string {
	return fmt.Sprintf("%v", *bo)
}

func (bo *BackendOptions) Set(value string) error {
	optionMap, err := parseBackendOptions(value)
	if err != nil {
		return err
	}
	*bo = append(*bo, optionMap)
	return nil
}

type Config struct {
	Host           string `env:"LOGINSRV_HOST"`
	Port           string `env:"LOGINSRV_PORT"`
	LogLevel       string `env:"LOGINSRV_LOG_LEVEL"`
	TextLogging    bool   `env:"LOGINSRV_TEXT_LOGGING"`
	JwtSecret      string `env:"LOGINSRV_JWT_SECRET"`
	SuccessUrl     string `env:"LOGINSRV_SUCCESS_URL"`
	CookieName     string `env:"LOGINSRV_COOKIE_NAME"`
	CookieHttpOnly bool   `env:"LOGINSRV_COOKIE_HTTP_ONLY"`
	Backends       BackendOptions
}

func DefaultConfig() *Config {
	return &Config{
		Host:           "localhost",
		Port:           "6789",
		LogLevel:       "info",
		JwtSecret:      randStringBytes(32),
		SuccessUrl:     "/",
		CookieName:     "jwt_token",
		CookieHttpOnly: true,
		Backends:       BackendOptions{},
	}
}
func ReadConfig() *Config {
	config := DefaultConfig()

	env.Parse(config)

	flag.StringVar(&config.Host, "host", config.Host, "The host to listen on")
	flag.StringVar(&config.Port, "port", config.Port, "The port to listen on")
	flag.StringVar(&config.LogLevel, "log-level", config.LogLevel, "The log level")
	flag.BoolVar(&config.TextLogging, "text-logging", config.TextLogging, "Log in text format instead of json")
	flag.StringVar(&config.JwtSecret, "jwt-secret", "random key", "The secret to sign the jwt token")
	flag.StringVar(&config.CookieName, "cookie-name", config.CookieName, "The name of the jwt cookie")
	flag.BoolVar(&config.CookieHttpOnly, "cookie-http-only", config.CookieHttpOnly, "Set the cookie with the http only flag")
	flag.StringVar(&config.SuccessUrl, "success-url", config.SuccessUrl, "The url to redirect after login")
	flag.Var(&config.Backends, "backend", "Backend configuration in form 'provider=name,key=val,key=...', can be declared multiple times")

	flag.Parse()
	return config
}

func parseBackendOptions(b string) (map[string]string, error) {
	opts := map[string]string{}
	pairs := strings.Split(b, ",")
	for _, p := range pairs {
		pair := strings.SplitN(p, "=", 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("provider configuration has to be in form 'provider=name,key1=value1,key2=..', but was %v", p)
		}
		opts[pair[0]] = pair[1]
	}
	if _, exist := opts["provider"]; !exist {
		return nil, fmt.Errorf("no provider name specified in %v", b)
	}
	return opts, nil
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
