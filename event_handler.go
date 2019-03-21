package main

import "time"
import "github.com/godbus/dbus"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/mousebind"
import "github.com/BurntSushi/xgbutil/keybind"
import "github.com/eugene-eeo/kaze/tctx"

type ActionType int
type Reason uint32

const (
	ActionShowAll = ActionType(iota)
	ActionCloseOne
	ActionCloseTop
	ActionContextMenu

	ReasonExpired = Reason(1 + iota)
	ReasonUserDismissed
	ReasonCloseNotification
	ReasonUndefined
)

type Action struct {
	Type   ActionType
	Target uint32
}

type UidPair struct {
	Uid          uint
	Notification *Notification
	ExpiryReq    uint
	PopupAgeReq  uint
	ActionReq    uint
}

type EventHandler struct {
	conn             *dbus.Conn
	nextUid          uint
	closeChan        chan uint32 // Used for HandleClose
	actionChan       chan Action
	notificationChan chan *Notification
	pairs            *CappedPairs
	display          *PopupDisplay
	// Passed to display
	contextMenuFunc func(*Notification)
	closeOneFunc    func(*Notification)
}

func NewEventHandler(conn *dbus.Conn) *EventHandler {
	X, err := xgbutil.NewConn()
	if err != nil {
		panic(err)
	}
	mousebind.Initialize(X)
	keybind.Initialize(X)
	var ev *EventHandler
	ev = &EventHandler{
		conn:             conn,
		nextUid:          0,
		closeChan:        make(chan uint32),
		actionChan:       make(chan Action),
		notificationChan: make(chan *Notification),
		pairs:            NewCappedPairs(conf.Core.Max),
		display:          &PopupDisplay{X, map[uint]*Popup{}},
		contextMenuFunc:  func(n *Notification) { ev.actionChan <- Action{Type: ActionContextMenu, Target: n.Id} },
		closeOneFunc:     func(n *Notification) { ev.actionChan <- Action{Type: ActionCloseOne, Target: n.Id} },
	}

	showAll := keybind.KeyPressFun(func(x *xgbutil.XUtil, _ xevent.KeyPressEvent) {
		ev.actionChan <- Action{Type: ActionShowAll}
	})
	closeTop := keybind.KeyPressFun(func(x *xgbutil.XUtil, _ xevent.KeyPressEvent) {
		ev.actionChan <- Action{Type: ActionCloseTop}
	})

	showAll.Connect(X, X.RootWin(), conf.Bindings.ShowAll, true)
	closeTop.Connect(X, X.RootWin(), conf.Bindings.CloseTop, true)

	go xevent.Main(X)
	go ev.Loop()
	return ev
}

func (e *EventHandler) EmitClosed(id uint32, reason Reason) {
	e.conn.Emit("/org/freedesktop/Notifications", "org.freedesktop.Notifications.NotificationClosed", id, reason)
}

func (e *EventHandler) EmitAction(id uint32, action_key string) {
	e.conn.Emit("/org/freedesktop/Notifications", "org.freedesktop.Notifications.ActionInvoked", id, action_key)
}

func (e *EventHandler) HandleNotification(n *Notification) {
	e.notificationChan <- n
}

func (e *EventHandler) HandleClose(id uint32) *dbus.Error {
	e.closeChan <- id
	if <-e.closeChan == 0 {
		return dbus.NewError("Close", []interface{}{})
	}
	return nil
}

func (e *EventHandler) deleteNotification(id uint32, uid uint, reason Reason) {
	if e.pairs.Get(id) != nil {
		e.pairs.Delete(id)
		e.display.Destroy(uid)
		e.EmitClosed(id, reason)
	}
}

func (e *EventHandler) handleExpiry(id uint) {
	found := false
	for _, p := range *e.pairs.lru {
		if p.ActionReq == id {
			e.deleteNotification(p.Notification.Id, p.Uid, ReasonUndefined)
			found = true
			break
		}
		if p.ExpiryReq == id {
			e.deleteNotification(p.Notification.Id, p.Uid, ReasonExpired)
			found = true
			break
		}
		if p.PopupAgeReq == id {
			e.display.Close(p.Uid)
			found = true
			break
		}
	}
	if found {
		e.draw()
	}
}

func (e *EventHandler) draw() {
	e.display.Draw(*e.pairs.lru)
}

func (e *EventHandler) Loop() {
	for {
		select {
		case n := <-e.notificationChan:
			// Set a max age if it's not a critical notification
			maxAge := conf.Core.MaxPopupAge.Duration
			maxTimeout := conf.Core.MaxAge.Duration
			if n.Hints.Urgency == UrgencyCritical {
				// Critical Notifications
				maxAge = -1
			} else {
				// Non-critical
				timeout := time.Duration(n.ExpireTimeout) * time.Millisecond
				if maxTimeout > timeout {
					maxTimeout = timeout
				}
			}
			e.nextUid++
			old := e.nextUid
			uid := e.nextUid
			if u := e.pairs.Get(n.Id); u != nil {
				// We have seen this before
				e.pairs.Delete(n.Id)
				old = u.Uid
			}
			// add expiries
			u := &UidPair{
				Uid:          uid,
				Notification: n,
				ExpiryReq:    tctx.Request(maxTimeout),
				PopupAgeReq:  tctx.Request(maxAge),
			}
			// show
			e.display.Show(old, uid, n, e.contextMenuFunc, e.closeOneFunc)
			// remove excess
			excess := e.pairs.Insert(n.Id, u)
			if excess != nil {
				e.deleteNotification(excess.Notification.Id, excess.Uid, ReasonUndefined)
			}
			e.draw()

		case id := <-tctx.Listen():
			e.handleExpiry(id)

		case a := <-e.actionChan:
			switch a.Type {
			case ActionShowAll:
				for _, u := range *e.pairs.lru {
					e.display.Show(u.Uid, u.Uid, u.Notification, e.contextMenuFunc, e.closeOneFunc)
					if u.Notification.Hints.Urgency != UrgencyCritical {
						u.PopupAgeReq = tctx.Request(conf.Core.MaxPopupAge.Duration)
					}
				}
				e.draw()
			case ActionCloseTop:
				if e.pairs.lru.Len() > 0 {
					pair := (*e.pairs.lru)[0]
					e.deleteNotification(pair.Notification.Id, pair.Uid, ReasonUserDismissed)
				}
			case ActionCloseOne:
				if u := e.pairs.Get(a.Target); u != nil {
					e.deleteNotification(a.Target, u.Uid, ReasonUserDismissed)
				}
			case ActionContextMenu:
				if u := e.pairs.Get(a.Target); u != nil {
					go execMixedSelector(u.Notification.Actions, u.Notification.Body.Hyperlinks, func(action_key string) {
						// Only emit if we have a valid action
						if len(action_key) > 0 {
							e.EmitAction(u.Notification.Id, action_key)
						}
						// Otherwise we have no events!
						if !u.Notification.Hints.Resident {
							u.ActionReq = tctx.Request(0)
						}
					})
				}
			}
		case id := <-e.closeChan:
			if t := e.pairs.Get(id); t != nil {
				e.deleteNotification(id, t.Uid, ReasonCloseNotification)
				e.closeChan <- id
			} else {
				e.closeChan <- 0
			}
		}
	}
}
