package caddy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	"github.com/tarent/loginsrv/logging"
	"github.com/tarent/loginsrv/login"

	// Import all backends, packaged with the caddy plugin
	_ "github.com/tarent/loginsrv/htpasswd"
	_ "github.com/tarent/loginsrv/httpupstream"
	_ "github.com/tarent/loginsrv/oauth2"
	_ "github.com/tarent/loginsrv/osiam"
	_ "github.com/tarent/loginsrv/saml"
	// Import SAML library
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
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

	secretProvidedByConfig := false
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
			return cfg, fmt.Errorf("Unknown parameter for login directive: %v (%v:%v)", name, c.File(), c.Line())
		}
		err := f.Value.Set(value)
		if err != nil {
			return cfg, fmt.Errorf("Invalid value for parameter %v: %v (%v:%v)", name, value, c.File(), c.Line())
		}

		if name == "jwt-secret" {
			secretProvidedByConfig = true
		}
	}

	if err := cfg.ResolveFileReferences(); err != nil {
		return nil, err
	}

	secretFromEnv, secretFromEnvWasSetBefore := os.LookupEnv("JWT_SECRET")
	if !secretProvidedByConfig && secretFromEnvWasSetBefore {
		cfg.JwtSecret = secretFromEnv
	}
	if !secretFromEnvWasSetBefore {
		// populate the secret to caddy.jwt,
		// but do not change a environment variable, which somebody has set it.
		os.Setenv("JWT_SECRET", cfg.JwtSecret)
	}

	if cfg.Azure.Enabled {

		if cfg.Azure.TenantID == "" {
			return cfg, fmt.Errorf("Azure AD Tenant ID not found")
		}

		logging.Logger.Infof("plugin/login/config: Azure AD Tenant ID => %s", cfg.Azure.TenantID)

		if cfg.Azure.ApplicationID == "" {
			return cfg, fmt.Errorf("Azure AD Application ID not found")
		}

		logging.Logger.Infof("plugin/login/config: Azure AD Application ID => %s", cfg.Azure.ApplicationID)

		if cfg.Azure.ApplicationName == "" {
			return cfg, fmt.Errorf("Azure AD Application Name not found")
		}

		logging.Logger.Infof("plugin/login/config: Azure AD Application Name => %s", cfg.Azure.ApplicationName)

		if cfg.Azure.IdpMetadataLocation == "" {
			cfg.Azure.IdpMetadataLocation = fmt.Sprintf(
				"https://login.microsoftonline.com/%s/federationmetadata/2007-06/federationmetadata.xml",
				cfg.Azure.TenantID,
			)
		}

		logging.Logger.Infof("plugin/login/config: Azure AD IdP Metadata Location => %s", cfg.Azure.IdpMetadataLocation)

		if cfg.Azure.IdpSignCertLocation == "" {
			return cfg, fmt.Errorf("Azure AD IdP Signing Certificate not found")
		}

		logging.Logger.Infof("plugin/login/config: Azure AD IdP Signing Certificate => %s", cfg.Azure.IdpSignCertLocation)

		idpSignCert, err := readCertFile(cfg.Azure.IdpSignCertLocation)
		if err != nil {
			return cfg, err
		}

		cfg.Azure.LoginURL = fmt.Sprintf(
			"https://account.activedirectory.windowsazure.com/applications/signin/%s/%s?tenantId=%s",
			cfg.Azure.ApplicationName, cfg.Azure.ApplicationID, cfg.Azure.TenantID,
		)

		logging.Logger.Infof("plugin/login/config: Azure AD Login URL => %s", cfg.Azure.LoginURL)

		azureOptions := samlsp.Options{}

		if strings.HasPrefix(cfg.Azure.IdpMetadataLocation, "http") {
			idpMetadataURL, err := url.Parse(cfg.Azure.IdpMetadataLocation)
			if err != nil {
				return cfg, err
			}
			cfg.Azure.IdpMetadataURL = idpMetadataURL
			azureOptions.URL = *idpMetadataURL
			idpMetadata, err := samlsp.FetchMetadata(
				context.Background(),
				http.DefaultClient,
				*idpMetadataURL,
			)
			if err != nil {
				return cfg, err
			}
			azureOptions.IDPMetadata = idpMetadata

		} else {
			metadataFileContent, err := ioutil.ReadFile(cfg.Azure.IdpMetadataLocation)
			if err != nil {
				return cfg, err
			}
			idpMetadata, err := samlsp.ParseMetadata(metadataFileContent)
			if err != nil {
				return cfg, err
			}
			azureOptions.IDPMetadata = idpMetadata
		}
		sp := samlsp.DefaultServiceProvider(azureOptions)
		sp.AllowIDPInitiated = true
		//sp.EntityID = sp.IDPMetadata.EntityID

		cfgAcsURL, _ := url.Parse(cfg.Azure.AcsURL)
		sp.AcsURL = *cfgAcsURL
		cfgMetadataURL, _ := url.Parse(cfg.Azure.MetadataURL)
		sp.MetadataURL = *cfgMetadataURL

		if cfg.Azure.IdpMetadataURL != nil {
			sp.MetadataURL = *cfg.Azure.IdpMetadataURL
		}

		for i := range sp.IDPMetadata.IDPSSODescriptors {
			idpSSODescriptor := &sp.IDPMetadata.IDPSSODescriptors[i]
			keyDescriptor := &saml.KeyDescriptor{
				Use: "signing",
				KeyInfo: saml.KeyInfo{
					XMLName: xml.Name{
						Space: "http://www.w3.org/2000/09/xmldsig#",
						Local: "KeyInfo",
					},
					Certificate: idpSignCert,
				},
			}
			idpSSODescriptor.KeyDescriptors = append(idpSSODescriptor.KeyDescriptors, *keyDescriptor)
			break
		}

		cfg.Azure.ServiceProvider = &sp
	}

	return cfg, nil
}

func readCertFile(filePath string) (string, error) {
	var buffer bytes.Buffer
	var RecordingEnabled bool
	fileHandle, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fileHandle.Close()

	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "-----") {
			if strings.Contains(line, "BEGIN CERTIFICATE") {
				RecordingEnabled = true
				continue
			}
			if strings.Contains(line, "END CERTIFICATE") {
				break
			}
		}
		if RecordingEnabled {
			buffer.WriteString(strings.TrimSpace(line))
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
