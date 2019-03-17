package x

import "github.com/godbus/dbus"
import "github.com/eugene-eeo/kaze/libkaze"

type XHandler struct {
}

func (_ *XHandler) Capabilities() []string {
	return []string{"body"}
}

func (_ *XHandler) HandleNotification(n *libkaze.Notification) {
}

func (_ *XHandler) HandleClose(id uint32) *dbus.Error {
	return nil
}

func (_ *XHandler) HandleTimeout(id uint32) {
}
