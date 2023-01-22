package utils

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port            int      `yaml:"lb_port"`
	RetryLimit      int      `yaml:"retry_limit"`
	MaxAttemptLimit int      `yaml:"max_attempt_limit"`
	Backends        []string `yaml:"backends"`
}

func GetLBConfig() (*Config, error) {
	var config Config
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}
	if len(config.Backends) == 0 {
		return nil, errors.New("backend hosts expected, none provided")
	}

	if config.Port == 0 {
		return nil, errors.New("load balancer port not found")
	}

	return &config, nil
}
