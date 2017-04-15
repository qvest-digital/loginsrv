package login

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env"
	"math/rand"
	"os"
	"strings"
	"time"
)

var DefaultConfig Config

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	DefaultConfig = Config{
		Host:           "localhost",
		Port:           "6789",
		LogLevel:       "info",
		JwtSecret:      randStringBytes(32),
		SuccessUrl:     "/",
		CookieName:     "jwt_token",
		CookieHttpOnly: true,
		Backends:       Options{},
		Oauth:          Options{},
	}
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
	Backends       Options
	Oauth          Options
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
	config := DefaultConfig

	err := env.Parse(&config)
	if err != nil {
		return nil, err
	}

	for _, v := range os.Environ() {
		pair := strings.SplitN(v, "=", 2)
		if len(pair) == 2 && strings.HasPrefix(pair[0], "LOGINSRV_BACKEND") {
			(&config.Backends).Set(pair[1])
		}
		if len(pair) == 2 && strings.HasPrefix(pair[0], "LOGINSRV_OAUTH") {
			(&config.Oauth).Set(pair[1])
		}
	}

	f.StringVar(&config.Host, "host", config.Host, "The host to listen on")
	f.StringVar(&config.Port, "port", config.Port, "The port to listen on")
	f.StringVar(&config.LogLevel, "log-level", config.LogLevel, "The log level")
	f.BoolVar(&config.TextLogging, "text-logging", config.TextLogging, "Log in text format instead of json")
	f.StringVar(&config.JwtSecret, "jwt-secret", "random key", "The secret to sign the jwt token")
	f.StringVar(&config.CookieName, "cookie-name", config.CookieName, "The name of the jwt cookie")
	f.BoolVar(&config.CookieHttpOnly, "cookie-http-only", config.CookieHttpOnly, "Set the cookie with the http only flag")
	f.StringVar(&config.SuccessUrl, "success-url", config.SuccessUrl, "The url to redirect after login")
	f.Var(&config.Backends, "backend", "Backend configuration in form 'provider=name,key=val,key=...', can be declared multiple times")
	f.Var(&config.Oauth, "oauth", "Oauth provider configuration in form 'provider=name,key=val,key=...', can be declared multiple times")

	err = f.Parse(args)
	if err != nil {
		return nil, err
	}

	if config.JwtSecret == "random key" {
		if s, set := os.LookupEnv("LOGINSRV_JWT_SECRET"); set {
			config.JwtSecret = s

		} else {
			config.JwtSecret = DefaultConfig.JwtSecret
		}
	}

	return &config, err
}

func parseOptions(b string) (map[string]string, error) {
	opts := map[string]string{}
	pairs := strings.Split(b, ",")
	for _, p := range pairs {
		pair := strings.SplitN(p, "=", 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("provider configuration has to be in form 'provider=name,key1=value1,key2=..', but was %v", p)
		}
		opts[pair[0]] = pair[1]
	}
	return opts, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type Options []map[string]string

func (bo *Options) String() string {
	return fmt.Sprintf("%v", *bo)
}

func (bo *Options) Set(value string) error {
	optionMap, err := parseOptions(value)
	if err != nil {
		return err
	}
	*bo = append(*bo, optionMap)
	return nil
}
