package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const DEFAULT_CONFIG_PATH = "config.yml"

type Config struct {
	Port    int  `yaml:"port"`
	IsDebug bool `yaml:"is_debug"`
}

func MustParse() *Config {
	return MustParseByPath(DEFAULT_CONFIG_PATH)
}

func MustParseByPath(cfgPath string) *Config {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		panic(fmt.Errorf("error while reading config file: %w", err))
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		panic(fmt.Errorf("error while unmarshaling config file: %w", err))
	}

	return &cfg
}
