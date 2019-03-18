package libkaze

import "time"
import "github.com/godbus/dbus"

type expiryPair struct {
	uid uint
	id  uint32
}

// HandlerWrapper wraps a NotificationHandler so that the underlying
// handler doesn't have to worry about timeouts and is called
// synchronously
type HandlerWrapper struct {
	uid                    uint
	conn                   *dbus.Conn
	timeouts               map[uint32]uint
	expiryChan             chan expiryPair
	errorsChan             chan *dbus.Error
	notificationChan       chan *Notification
	notificationClosedChan chan uint32
	closedChan             chan uint32
	handler                NotificationHandler
}

func WrapHandler(conn *dbus.Conn, n NotificationHandler) *HandlerWrapper {
	return &HandlerWrapper{
		conn:                   conn,
		timeouts:               map[uint32]uint{},
		errorsChan:             make(chan *dbus.Error),
		notificationChan:       make(chan *Notification),
		notificationClosedChan: make(chan uint32),
		expiryChan:             make(chan expiryPair),
		closedChan:             make(chan uint32),
		handler:                n,
	}
}

func (h *HandlerWrapper) Capabilities() []string {
	return h.handler.Capabilities()
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
// the notification dialog. This emits the NotificationClosed signal appropriately.
func (h *HandlerWrapper) SilentNotificationClose(id uint32) {
	h.closedChan <- id
}

func (h *HandlerWrapper) Loop() {
	for {
		select {
		case n := <-h.notificationChan:
			h.uid++
			uid := h.uid
			// Urgent Notifications should not have a timeout
			if n.ExpireTimeout != -1 && n.Hints.Urgency != UrgencyCritical {
				// Associate with each notification a uid, that way we can check
				// if a notification has expired correctly
				go func() {
					time.Sleep(time.Millisecond * time.Duration(n.ExpireTimeout))
					h.expiryChan <- expiryPair{uid, n.Id}
				}()
			}
			h.timeouts[n.Id] = uid
			h.handler.HandleNotification(n)

		case id := <-h.notificationClosedChan:
			if _, ok := h.timeouts[id]; ok {
				delete(h.timeouts, id)
				err := h.handler.HandleClose(id)
				if err == nil {
					h.emitNotificationClosed(id)
				}
				h.errorsChan <- err
			} else {
				h.errorsChan <- &dbus.Error{}
			}

		case pair := <-h.expiryChan:
			if uid, ok := h.timeouts[pair.id]; ok && uid == pair.uid {
				delete(h.timeouts, pair.id)
				h.handler.HandleTimeout(pair.id)
				h.emitNotificationClosed(pair.id)
			}

		case id := <-h.closedChan:
			if _, ok := h.timeouts[id]; ok {
				delete(h.timeouts, id)
				h.emitNotificationClosed(id)
			}
		}
	}
}
