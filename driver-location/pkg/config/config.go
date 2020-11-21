package config

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	App *struct {
		DriverLocationsLimit int `yaml:"driver_locations_limit"`
	} `yaml:"app"`

	Redis *struct {
		Address string `yaml:"address"`
	} `yaml:"redis"`

	HTTPServer *struct {
		Port            int           `yaml:"port"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	} `yaml:"http_server"`

	NSQ *struct {
		Topic            string   `yaml:"topic"`
		Channel          string   `yaml:"channel"`
		DaemonAddresses  []string `yaml:"daemon_addresses"`
		WorkersNum       int      `yaml:"workers_num"`
	} `yaml:"nsq"`
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
