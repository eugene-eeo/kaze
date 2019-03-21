package main

import "time"
import "github.com/godbus/dbus"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/mousebind"
import "github.com/BurntSushi/xgbutil/keybind"
import "github.com/eugene-eeo/kaze/tctx"

type ExpiryType int
type ActionType int

const (
	ActionShowAll = ActionType(iota)
	ActionCloseOne
	ActionCloseLatest
	ActionContextMenu

	ExpiryTimeout = ExpiryType(iota)
	ExpiryAction
	ExpiryPopup
)

type Expiry struct {
	Type   ExpiryType
	Id     uint32
	Target uint
}

type Action struct {
	Type   ActionType
	Target uint32
}

type UidPair struct {
	Uid          uint
	Notification *Notification
}

type EventHandler struct {
	conn             *dbus.Conn
	nextUid          uint
	closeChan        chan uint32 // Used for HandleClose
	closedChan       chan bool   // Used for HandleClose
	actionChan       chan Action
	notificationChan chan *Notification
	expiries         map[uint]Expiry
	pairs            *CappedPairs
	display          *PopupDisplay
}

func NewEventHandler(conn *dbus.Conn) *EventHandler {
	X, err := xgbutil.NewConn()
	if err != nil {
		panic(err)
	}
	mousebind.Initialize(X)
	keybind.Initialize(X)
	ev := EventHandler{
		conn:             conn,
		nextUid:          0,
		closeChan:        make(chan uint32),
		closedChan:       make(chan bool),
		actionChan:       make(chan Action),
		notificationChan: make(chan *Notification),
		expiries:         map[uint]Expiry{},
		pairs:            NewCappedPairs(conf.Core.Max),
		display:          &PopupDisplay{X, map[uint]*Popup{}},
	}

	showAll := keybind.KeyPressFun(func(x *xgbutil.XUtil, _ xevent.KeyPressEvent) {
		ev.actionChan <- Action{Type: ActionShowAll}
	})
	closeLatest := keybind.KeyPressFun(func(x *xgbutil.XUtil, _ xevent.KeyPressEvent) {
		ev.actionChan <- Action{Type: ActionCloseLatest}
	})

	showAll.Connect(X, X.RootWin(), conf.Bindings.ShowAll, true)
	closeLatest.Connect(X, X.RootWin(), conf.Bindings.CloseLatest, true)

	go xevent.Main(X)
	go ev.Loop()
	return &ev
}

func (e *EventHandler) EmitClosed(id uint32, reason uint32) {
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
	if !<-e.closedChan {
		return dbus.NewError("Close", []interface{}{})
	}
	return nil
}

func (e *EventHandler) GetContextMenuFunc() func(*Notification) {
	return func(n *Notification) {
		e.actionChan <- Action{Type: ActionContextMenu, Target: n.Id}
	}
}

func (e *EventHandler) GetCloseFunc() func(*Notification) {
	return func(n *Notification) {
		e.actionChan <- Action{Type: ActionCloseOne, Target: n.Id}
	}
}

func (e *EventHandler) deleteNotification(id uint32, uid uint, reason uint32) {
	if e.pairs.Get(id) != nil {
		e.pairs.Delete(id)
		e.display.Destroy(uid)
		e.EmitClosed(id, reason)
	}
}

func (e *EventHandler) handleExpiry(exp Expiry) {
	t := e.pairs.Get(exp.Id)
	if t != nil && t.Uid == exp.Target {
		switch exp.Type {
		case ExpiryAction:
			e.deleteNotification(exp.Id, exp.Target, 4)
		case ExpiryTimeout:
			e.deleteNotification(exp.Id, exp.Target, 1)
		case ExpiryPopup:
			e.display.Close(exp.Target)
		}
		e.display.Draw()
	}
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
				old = u.Uid
			}
			// add expiries
			e.expiries[tctx.Request(maxAge)] = Expiry{ExpiryPopup, n.Id, uid}
			e.expiries[tctx.Request(maxTimeout)] = Expiry{ExpiryTimeout, n.Id, uid}
			// show
			e.display.Show(old, uid, n, e.GetContextMenuFunc(), e.GetCloseFunc())
			// remove excess
			excess := e.pairs.Insert(n.Id, &UidPair{uid, n})
			if excess != nil {
				e.deleteNotification(excess.Notification.Id, excess.Uid, 4)
			}
			e.display.Draw()

		case id := <-tctx.Listen():
			if _, ok := e.expiries[id]; ok {
				e.handleExpiry(e.expiries[id])
				delete(e.expiries, id)
			}

		case a := <-e.actionChan:
			switch a.Type {
			case ActionShowAll:
				for _, u := range e.pairs.pairs {
					e.display.Show(u.Uid, u.Uid, u.Notification, e.GetContextMenuFunc(), e.GetCloseFunc())
					if u.Notification.Hints.Urgency != UrgencyCritical {
						e.expiries[tctx.Request(conf.Core.MaxPopupAge.Duration)] = Expiry{
							Type:   ExpiryPopup,
							Id:     u.Notification.Id,
							Target: u.Uid,
						}
					}
				}
				e.display.Draw()
			case ActionCloseLatest:
				target := uint32(0)
				uid := uint(0)
				for id, u := range e.pairs.pairs {
					if u.Uid > uid {
						uid = u.Uid
						target = id
					}
				}
				if target != 0 {
					e.deleteNotification(target, uid, 2)
				}
			case ActionCloseOne:
				if u := e.pairs.Get(a.Target); u != nil {
					e.deleteNotification(a.Target, u.Uid, 2)
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
							e.expiries[tctx.Request(0)] = Expiry{ExpiryAction, a.Target, u.Uid}
						}
					})
				}
			}
		case id := <-e.closeChan:
			if t := e.pairs.Get(id); t != nil {
				e.deleteNotification(id, t.Uid, 3)
				e.closedChan <- true
			} else {
				e.closedChan <- false
			}
		}
	}
}
