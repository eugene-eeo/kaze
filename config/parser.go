package config

import "github.com/BurntSushi/toml"

func ConfigFromFile(path string) (*Config, error) {
	config := &Config{}
	_, err := toml.DecodeFile(path, config)
	return config, err
}
