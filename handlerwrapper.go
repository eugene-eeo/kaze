package main

import "time"
import "github.com/godbus/dbus"

// HandlerWrapper wraps a NotificationHandler so that the underlying
// handler doesn't have to worry about timeouts and is called
// synchronously
type HandlerWrapper struct {
	conn                   *dbus.Conn
	open                   map[uint32]bool
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
		open:                   make(map[uint32]bool),
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
			h.open[n.Id] = true
			h.handler.HandleNotification(n)
		case id := <-h.expiryChan:
			if h.open[id] {
				h.handler.HandleTimeout(id)
				h.emitNotificationClosed(id)
				delete(h.open, id)
			}
		case id := <-h.notificationClosedChan:
			if h.open[id] {
				delete(h.open, id)
				err := h.handler.HandleClose(id)
				if err == nil {
					h.emitNotificationClosed(id)
				}
				h.errorsChan <- err
			} else {
				h.errorsChan <- &dbus.Error{}
			}
		// hack, just mark the id as closed
		// should be used when for instance the user closes the window
		case id := <-h.closedChan:
			delete(h.open, id)
		}
	}
}
