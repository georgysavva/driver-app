package config

import (
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/heetch/georgysavva-technical-test/gateway/pkg/gateway"
)

type Config struct {
	URLs []*gateway.Endpoint `yaml:"urls"`

	HTTPServer *struct {
		Port            int           `yaml:"port"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	} `yaml:"http_server"`

	NSQ *struct {
		DaemonAddress string `yaml:"daemon_address"`
	} `yaml:"nsq"`
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
