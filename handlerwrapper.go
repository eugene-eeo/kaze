package main

import "time"
import "github.com/godbus/dbus"
import "golang.org/x/tools/container/intsets"

// HandlerWrapper wraps a NotificationHandler so that the underlying
// handler doesn't have to worry about timeouts and is called
// synchronously
type HandlerWrapper struct {
	conn                   *dbus.Conn
	open                   intsets.Sparse
	errorsChan             chan *dbus.Error
	notificationChan       chan *Notification
	notificationClosedChan chan uint32
	closedChan             chan uint32
	expiryChan             chan uint32
	handler                NotificationHandler
}

func WrapHandler(conn *dbus.Conn, n NotificationHandler) *HandlerWrapper {
	return &HandlerWrapper{
		conn:                   conn,
		open:                   intsets.Sparse{},
		errorsChan:             make(chan *dbus.Error),
		notificationChan:       make(chan *Notification),
		notificationClosedChan: make(chan uint32),
		closedChan:             make(chan uint32),
		expiryChan:             make(chan uint32),
		handler:                n,
	}
}

func (h *HandlerWrapper) HandleNotification(n *Notification) {
	h.notificationChan <- n
}

func (h *HandlerWrapper) HandleClose(id uint32) *dbus.Error {
	h.notificationClosedChan <- id
	return <-h.errorsChan
}

func (h *HandlerWrapper) HandleTimeout(id uint32) {
}

func (h *HandlerWrapper) emitNotificationClosed(id uint32) {
	h.conn.Emit("/org/freedesktop/Notifications", "org.freedesktop.Notifications.NotificationClosed", id)
}

// SilentNotificationClose should be used to indicate that the underlying handler
// has closed the notification without intervention from dbus, e.g. the user closes
// the notification dialog.
func (h *HandlerWrapper) SilentNotificationClose(id uint32) {
	h.closedChan <- id
}

func (h *HandlerWrapper) Loop() {
	for {
		select {
		case n := <-h.notificationChan:
			if n.ExpireTimeout > 0 {
				go func() {
					time.Sleep(time.Millisecond * time.Duration(n.ExpireTimeout))
					h.expiryChan <- n.Id
				}()
			}
			h.open.Insert(int(n.Id))
			h.handler.HandleNotification(n)
		case id := <-h.expiryChan:
			x := int(id)
			if h.open.Has(x) {
				h.open.Remove(x)
				h.handler.HandleTimeout(id)
				h.emitNotificationClosed(id)
			}
		case id := <-h.notificationClosedChan:
			x := int(id)
			if h.open.Has(x) {
				h.open.Remove(x)
				err := h.handler.HandleClose(id)
				if err == nil {
					h.emitNotificationClosed(id)
				}
				h.errorsChan <- err
			} else {
				h.errorsChan <- &dbus.Error{}
			}
		case id := <-h.closedChan:
			h.open.Remove(int(id))
		}
	}
}
