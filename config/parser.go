package config

import "errors"
import "github.com/BurntSushi/toml"

var contextMenuNoCommand = errors.New("core.context_menu: no command given")
var linkOpenNoCommand = errors.New("core.link_opener: no command given")

func ConfigFromFile(path string) (*Config, error) {
	config := &Config{}
	_, err := toml.DecodeFile(path, config)
	if len(config.Core.ContextMenuProgram) == 0 {
		return nil, contextMenuNoCommand
	}
	if len(config.Core.LinkOpenProgram) == 0 {
		return nil, linkOpenNoCommand
	}
	return config, err
}
