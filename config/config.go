package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Port  string `yaml:"port"`
	Token struct {
		TtlAccess  int `yaml:"ttl_access"`
		TtlRefresh int `yaml:"ttl_refresh"`
	} `yaml:"token"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
