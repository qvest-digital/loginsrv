package caddy

import (
	"fmt"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/tarent/lib-compose/logging"
	"github.com/tarent/loginsrv/login"
	"os"
	"strconv"
)

func init() {
	caddy.RegisterPlugin("loginsrv", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
	httpserver.RegisterDevDirective("loginsrv", "jwt")
}

// setup configures a new loginsrv instance.
func setup(c *caddy.Controller) error {
	logging.Set("info", true)

	for c.Next() {
		args := c.RemainingArgs()

		if len(args) < 1 {
			return fmt.Errorf("Missing path argument for loginsrv directive (%v:%v)", c.File(), c.Line())
		}

		if len(args) > 1 {
			return fmt.Errorf("To many arguments for loginsrv directive %q (%v:%v)", args, c.File(), c.Line())
		}

		config, err := parseConfig(c)
		if err != nil {
			return err
		}

		if e, isset := os.LookupEnv("JWT_SECRET"); isset {
			config.JwtSecret = e
		} else {
			os.Setenv("JWT_SECRET", config.JwtSecret)
		}

		loginHandler, err := login.NewHandler(&config)
		if err != nil {
			return err
		}

		httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
			return NewCaddyHandler(next, args[0], loginHandler, &config)
		})
	}

	return nil
}

func parseConfig(c *caddy.Controller) (login.Config, error) {
	cfg := login.DefaultConfig
	cfg.Host = ""
	cfg.Port = ""
	cfg.LogLevel = ""

	for c.NextBlock() {
		name := c.Val()
		args := c.RemainingArgs()
		if len(args) != 1 {
			return cfg, fmt.Errorf("Wrong number of arguments for %v: %v (%v:%v)", name, args, c.File(), c.Line())
		}
		value := args[0]

		switch name {
		case "success-url":
			cfg.SuccessUrl = value
		case "cookie-name":
			cfg.CookieName = value
		case "cookie-http-only":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return cfg, fmt.Errorf("error parsing bool value %v: %v (%v:%v)", name, value, c.File(), c.Line())
			}
			cfg.CookieHttpOnly = b
		case "backend":
			err := (&cfg.Backends).Set(value)
			if err != nil {
				return cfg, fmt.Errorf("error parsing backend configuration %v: %v (%v:%v)", name, value, c.File(), c.Line())
			}
		default:
			return cfg, fmt.Errorf("Unknown option within loginsrv: %v (%v:%v)", name, c.File(), c.Line())
		}
	}

	return cfg, nil
}
