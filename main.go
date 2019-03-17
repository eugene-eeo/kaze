package main

import "github.com/godbus/dbus"

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
	handler := WrapHandler(conn, &NullHandler{conn})
	s := Service{
		id:      0,
		conn:    conn,
		handler: handler,
	}
	go handler.Loop()
	err = conn.Export(&s, "/org/freedesktop/Notifications", "org.freedesktop.Notifications")
	if err != nil {
		panic(err)
	}
	select {}
}
