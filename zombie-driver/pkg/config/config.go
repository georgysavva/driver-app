package config

import (
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	App struct {
		ZombiePredicate struct {
			DistanceThreshold int           `yaml:"distance_threshold"`
			TimeInterval      time.Duration `yaml:"time_interval"`
		}
	} `yaml:"app"`

	DriverLocationService struct {
		URL string `yaml:"url"`
	} `yaml:"driver_location_service"`

	HTTPServer struct {
		Port            int           `yaml:"port"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	} `yaml:"http_server"`
}

func ParseConfig(configPath string) (*Config, error) {
	fileContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file content")
	}

	conf := &Config{}
	if err := yaml.Unmarshal(fileContent, conf); err != nil {
		return nil, errors.Wrap(err, "failed to parse yaml content into Config struct")
	}

	return conf, nil
}
