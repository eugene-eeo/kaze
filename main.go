package main

import "github.com/godbus/dbus"
import "github.com/eugene-eeo/kaze/libkaze"
import kaze_x "github.com/eugene-eeo/kaze/x"

func main() {
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
	handler := libkaze.WrapHandler(conn, kaze_x.NewXHandler())
	service := libkaze.NewService(conn, handler)
	go handler.Loop()
	err = conn.Export(service, "/org/freedesktop/Notifications", "org.freedesktop.Notifications")
	if err != nil {
		panic(err)
	}
	select {}
}
