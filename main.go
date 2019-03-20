package main

import "github.com/godbus/dbus"
import "github.com/eugene-eeo/kaze/config"
import "github.com/BurntSushi/freetype-go/freetype/truetype"

var conf *config.Config
var fontBold *truetype.Font
var fontRegular *truetype.Font

func main() {
	var err error
	conf, err = config.ConfigFromFile("kaze.toml")
	if err != nil {
		panic(err)
	}

	// Parse fonts
	fontBold = mustReadFont(conf.Style.FontBold)
	fontRegular = mustReadFont(conf.Style.FontRegular)

	// DBus handshake
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	reply, err := conn.RequestName("org.freedesktop.Notifications", dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		panic("Name already taken")
	}
	service := NewService(conn, NewEventHandler(conn))
	err = conn.Export(service, "/org/freedesktop/Notifications", "org.freedesktop.Notifications")
	if err != nil {
		panic(err)
	}
	select {}
}
