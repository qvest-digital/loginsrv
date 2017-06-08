package caddy

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/tarent/loginsrv/logging"
	"github.com/tarent/loginsrv/login"

	// Import all backends, packaged with the caddy plugin
	_ "github.com/tarent/loginsrv/htpasswd"
	_ "github.com/tarent/loginsrv/httpupstream"
	_ "github.com/tarent/loginsrv/oauth2"
	_ "github.com/tarent/loginsrv/osiam"
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

		config, err := parseConfig(c)
		if err != nil {
			return err
		}

		if config.Template != "" && !filepath.IsAbs(config.Template) {
			config.Template = filepath.Join(httpserver.GetConfig(c).Root, config.Template)
		}

		if len(args) == 1 {
			logging.Logger.Warnf("DEPRECATED: Please set the login path by parameter login_path and not as directive argument (%v:%v)", c.File(), c.Line())
			config.LoginPath = path.Join(args[0], "/login")
		}

		if e, isset := os.LookupEnv("JWT_SECRET"); isset {
			config.JwtSecret = e
		} else {
			os.Setenv("JWT_SECRET", config.JwtSecret)
		}

		loginHandler, err := login.NewHandler(config)
		if err != nil {
			return err
		}

		httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
			return NewCaddyHandler(next, loginHandler, config)
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
		// caddy prefers '_' in parameter names,
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
			return cfg, c.ArgErr()
		}
		err := f.Value.Set(value)
		if err != nil {
			return cfg, c.Err(err.Error())
		}
	}

	return cfg, nil
}
