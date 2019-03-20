package config

import "github.com/BurntSushi/toml"

func ConfigFromFile(path string) (*Config, error) {
	config := &Config{}
	_, err := toml.DecodeFile(path, config)
	all := []string{"critical", "normal", "low"}
	base := config.Styles["base"]
	for _, profile := range all {
		style := config.Styles[profile]
		if style == nil {
			config.Styles[profile] = base
			continue
		}
		if style.BorderWidth == nil {
			style.BorderWidth = base.BorderWidth
		}
		if style.BorderColor == nil {
			style.BorderColor = base.BorderColor
		}
		if style.Padding == nil {
			style.Padding = base.Padding
		}
		if style.Foreground == nil {
			style.Foreground = base.Foreground
		}
		if style.Background == nil {
			style.Background = base.Background
		}
	}
	return config, err
}
