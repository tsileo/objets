package objets

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var (
	DefautRootPath = "./objets_data"
)

type Config struct {
	// Where objects will be stored
	UserDataDir string `yaml:"data_dir"`

	// Server listen
	UserListen string `yaml:"listen"`

	// TLS related config
	AutoTLS bool     `yaml:"tls_auto"`
	Domains []string `yaml:"tls_domains"`

	// Auth
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
}

func (c *Config) DataDir() string {
	if c.UserDataDir == "" {
		return DefautRootPath
	}
	return c.UserDataDir
}

func (c *Config) Listen() string {
	if c.UserListen != "" {
		return c.UserListen
	}
	if c.AutoTLS {
		return ":443"
	}
	return ":8060"
}

func newConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	if err := yaml.Unmarshal([]byte(data), &conf); err != nil {
		return nil, err
	}
	return conf, nil
}
