package main

import "github.com/godbus/dbus"
import "github.com/eugene-eeo/kaze/config"

var conf *config.Config

func main() {
	c, err := config.ConfigFromFile("kaze.toml")
	conf = c
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
	handler := NewXHandler()
	wrapper := WrapHandler(conn, handler)
	service := NewService(conn, wrapper)
	handler.Wrapper = wrapper
	go wrapper.Loop()
	go handler.Loop()
	err = conn.Export(service, "/org/freedesktop/Notifications", "org.freedesktop.Notifications")
	if err != nil {
		panic(err)
	}
	select {}
}
