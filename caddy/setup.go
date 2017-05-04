package caddy

import (
	"flag"
	"fmt"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/tarent/lib-compose/logging"
	_ "github.com/tarent/loginsrv/htpasswd"
	"github.com/tarent/loginsrv/login"
	_ "github.com/tarent/loginsrv/oauth2"
	_ "github.com/tarent/loginsrv/osiam"
	"os"
	"strings"
)

func init() {
	caddy.RegisterPlugin("login", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
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
		fmt.Printf("config %+v\n", config)
		loginHandler, err := login.NewHandler(config)
		if err != nil {
			return err
		}

		httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
			return NewCaddyHandler(next, args[0], loginHandler, config)
		})
	}

	return nil
}

func parseConfig(c *caddy.Controller) (*login.Config, error) {
	cfg := login.DefaultConfig()
	cfg.Host = ""
	cfg.Port = ""
	cfg.LogLevel = ""

	fs := flag.NewFlagSet("loginsrv-config", flag.ContinueOnError)
	cfg.ConfigureFlagSet(fs)

	for c.NextBlock() {
		// caddy preferes '_' in parameter names,
		// so we map them to the '-' from the command line flags
		// the replacement supports both, for backwards compatibility
		name := strings.Replace(c.Val(), "_", "-", -1)
		args := c.RemainingArgs()
		if len(args) != 1 {
			return cfg, fmt.Errorf("Wrong number of arguments for %v: %v (%v:%v)", name, args, c.File(), c.Line())
		}
		value := args[0]

		f := fs.Lookup(name)
		if f == nil {
			c.ArgErr()
			continue
		}
		f.Value.Set(value)
	}

	return cfg, nil
}
