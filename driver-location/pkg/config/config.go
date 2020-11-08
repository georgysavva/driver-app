package config

import (
	"io/ioutil"
	"path"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Redis struct {
		Address string `yaml:"address"`
	} `yaml:"redis"`

	App struct {
		DriverLocationsLimit int `yaml:"driver_locations_limit"`
	} `yaml:"app"`

	HTTPServer struct {
		Port            int           `yaml:"port"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	} `yaml:"http_server"`
}

const defaultConfigPath = "config.yaml"

func ParseConfig() (*Config, error) {
	// Improvement: allow to pass custom config path.

	currentFilePath, err := getCurrentFilePath()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	configAbsPath := path.Join(path.Dir(path.Dir(path.Dir(currentFilePath))), defaultConfigPath)

	fileContent, err := ioutil.ReadFile(configAbsPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file content")
	}
	conf := &Config{}
	if err := yaml.Unmarshal(fileContent, conf); err != nil {
		return nil, errors.Wrap(err, "failed to parse yaml content into Config struct")
	}

	return conf, nil
}

func getCurrentFilePath() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("failed to get current file path from runtime")
	}
	return filename, nil
}
