package config

import "errors"
import "time"
import "regexp"
import "encoding/hex"
import "github.com/BurntSushi/xgbutil/xgraphics"

var colorRegexp = regexp.MustCompile("^#[a-fA-F0-9]{6}$")
var colorMatchError = errors.New("cannot match color string")

type Config struct {
	Core     coreConfig     `toml:"core"`
	Style    styleConfig    `toml:"style"`
	Bindings bindingsConfig `toml:"bindings"`
}

type coreConfig struct {
	ContextMenuProgram []string `toml:"context_menu"`
	LinkOpenProgram    []string `toml:"link_opener"`
	MaxAge             duration `toml:"maxage"`
	MaxPopupAge        duration `toml:"maxpopupage"`
}

type styleConfig struct {
	XOffset     int   `toml:"x_offset"`
	YOffset     int   `toml:"y_offset"`
	Width       int   `toml:"width"`
	FontSize    int   `toml:"font_size"`
	BorderWidth int   `toml:"border_width"`
	BorderColor color `toml:"border_color"`
	Padding     int   `toml:"padding"`
	Fg          color `toml:"fg"`
	CriticalBg  color `toml:"critical_bg"`
	NormalBg    color `toml:"normal_bg"`
	LowBg       color `toml:"low_bg"`
}

type bindingsConfig struct {
	Filter      string `toml:"mouse_filter"`
	CloseOne    string `toml:"mouse_close_one"`
	CloseLatest string `toml:"kbd_close_latest"`
	ShowAll     string `toml:"kbd_show_all"`
}

type color struct {
	xgraphics.BGRA
}

func (c *color) UnmarshalText(text []byte) error {
	if !colorRegexp.Match(text) {
		return colorMatchError
	}
	buff, _ := hex.DecodeString(string(text)[1:])
	c.R = buff[0]
	c.G = buff[1]
	c.B = buff[2]
	c.A = 0xff
	return nil
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
