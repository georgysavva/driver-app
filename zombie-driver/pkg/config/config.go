package config

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/georgysavva/driver-app/zombie-driver/pkg/zombiedriver"
)

type Config struct {
	App struct {
		ZombiePredicate *zombiedriver.ZombiePredicate `yaml:"zombie_predicate"`
	} `yaml:"app"`

	DriverLocationService *struct {
		BaseURL string `yaml:"base_url"`
	} `yaml:"driver_location_service"`

	HTTPServer *struct {
		Port            int           `yaml:"port"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	} `yaml:"http_server"`
}

func ParseConfig(configPath string) (*Config, error) {
	fileContent, err := ioutil.ReadFile(filepath.Clean(configPath))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file content")
	}

	conf := &Config{}
	if err := yaml.Unmarshal(fileContent, conf); err != nil {
		return nil, errors.Wrap(err, "failed to parse yaml content into Config struct")
	}

	return conf, nil
}
