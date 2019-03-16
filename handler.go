package main

import "fmt"
import "github.com/godbus/dbus"

type NotificationHandler interface {
	HandleNotification(n *Notification)
	HandleClose(id uint32, conn *dbus.Conn)
}

type NullHandler struct{}

func (_ *NullHandler) HandleNotification(n *Notification) {
	fmt.Println(n)
}

func (_ *NullHandler) HandleClose(id uint32, conn *dbus.Conn) {
	conn.Emit("/org/freedesktop/Notifications", "org.freedesktop.Notifications.NotificationClosed", id)
}
