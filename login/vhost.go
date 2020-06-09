package login

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/tarent/loginsrv/oauth2"
	yaml "gopkg.in/yaml.v3"
)

// OAuthConfig contains the parsed oauth configuration
type OAuthConfig struct {
	Provider     string `yaml:"provider"`
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
	RedirectURI  string `yaml:"redirect-uri"`
	Score        string `yaml:"scope"`
}

// Convert options to a map
func (o *OAuthConfig) Convert() map[string]string {
	values := make(map[string]string)
	typ := reflect.TypeOf(*o)
	vtype := reflect.ValueOf(*o)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		k := strings.Split(field.Tag.Get("yaml"), ",")[0]
		if k == "provider" {
			continue
		}
		k = strings.Replace(k, "-", "_", -1)
		v := vtype.Field(i).String()
		if v == "" {
			continue
		}
		values[k] = v
	}
	return values
}

// VHostConfig contains per virtual host configuration
type VHostConfig struct {
	// Configuration object name (mandatory)
	Name string `yaml:"name"`
	// Hostname of the virtual host (defaults to Name)
	Hostname string `yaml:"hostname"`

	CookieDomain string `yaml:"cookie-domain"`
	Template     string `yaml:"template"`

	BackendConfigs []yaml.Node     `yaml:"backends"`
	OAuthConfigs   []OAuthConfig   `yaml:"oauth"`
	Users          []userFileEntry `yaml:"users"`
}

// VirtualHost contains the state for a virtual host.
type VirtualHost struct {
	Config   VHostConfig
	Backends []Backend
}

func (v *VirtualHost) makeHandler(globalConfig *Config) (*Handler, error) {
	claims := &userClaimsFile{
		userFileEntries: v.Config.Users,
	}
	vhostOauth := oauth2.NewManager()
	for _, authConfig := range v.Config.OAuthConfigs {
		opts := authConfig.Convert()
		if err := vhostOauth.AddConfig(authConfig.Provider, opts); err != nil {
			return nil, err
		}
	}

	config := *globalConfig
	if v.Config.CookieDomain != "" {
		config.CookieDomain = v.Config.CookieDomain
	}
	if v.Config.Template != "" {
		config.Template = v.Config.Template
	}

	return &Handler{
		backends:   v.Backends,
		config:     &config,
		userClaims: claims.Claims,
		oauth:      vhostOauth,
	}, nil
}

func decodeBackend(node *yaml.Node) (Backend, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("Unexpected yaml object type for backend")
	}

	opts := make(map[string]string, len(node.Content)/2)
	var key string
	for i, subnode := range node.Content {
		if i&1 == 0 {
			key = subnode.Value
			continue
		}
		opts[key] = subnode.Value
	}

	p, exists := GetProvider(opts["name"])
	if !exists {
		return nil, fmt.Errorf("Undefined backend: %s", opts["name"])
	}
	delete(opts, "name")
	return p(opts)
}

// UnmarshalYAML decodes the vhost config object
func (v *VirtualHost) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&v.Config); err != nil {
		return err
	}
	if v.Config.Name == "" {
		return fmt.Errorf("vhost object must define name attribute")
	}

	for _, node := range v.Config.BackendConfigs {
		backend, err := decodeBackend(&node)
		if err != nil {
			return err
		}
		v.Backends = append(v.Backends, backend)
	}
	for _, oauthConfig := range v.Config.OAuthConfigs {
		if _, exists := oauth2.GetProvider(oauthConfig.Provider); !exists {
			return fmt.Errorf("Unknown oauth2 provider: %s", oauthConfig.Provider)
		}
	}

	if v.Config.Hostname == "" {
		v.Config.Hostname = v.Config.Name
	}
	return nil
}
