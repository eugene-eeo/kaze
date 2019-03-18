package main

import "fmt"
import "github.com/godbus/dbus"

type NotificationHandler interface {
	Capabilities() []string
	HandleNotification(n *Notification)
	HandleClose(id uint32) *dbus.Error
	HandleTimeout(id uint32)
}

// Just for debugging
type NullHandler struct{}

func (_ *NullHandler) Capabilities() []string {
	return []string{"body"}
}

func (_ *NullHandler) HandleNotification(n *Notification) {
	fmt.Println(n)
}

func (_ *NullHandler) HandleClose(id uint32) *dbus.Error {
	return nil
}

func (_ *NullHandler) HandleTimeout(id uint32) {
	fmt.Println(id, "timeout")
}
