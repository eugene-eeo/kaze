package main

import "github.com/godbus/dbus"
import "github.com/eugene-eeo/kaze/config"

var conf *config.Config

func main() {
	var err error
	conf, err = config.ConfigFromFile("kaze.toml")
	if err != nil {
		panic(err)
	}
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
