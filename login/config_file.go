package login

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v3"
)

// yamlConfig holds the top level configuration data.
type yamlConfig struct {
	LoginPath       string        `yaml:"login-path"`
	HealthCheckPath string        `yaml:"health-check-path"`
	VHosts          []VirtualHost `yaml:"vhosts"`
}

func parseConfigData(config *Config, data []byte) error {
	var configdata yamlConfig
	if err := yaml.Unmarshal(data, &configdata); err != nil {
		return err
	}

	if configdata.LoginPath != "" {
		config.LoginPath = configdata.LoginPath
	}
	if configdata.HealthCheckPath != "" {
		config.HealthCheckPath = configdata.HealthCheckPath
	}
	config.VHosts = configdata.VHosts
	return nil
}

func parseConfigFile(config *Config) error {
	if config.ConfigFile == "" {
		return nil
	}
	b, err := ioutil.ReadFile(config.ConfigFile)
	if err != nil {
		return err
	}
	return parseConfigData(config, b)
}
