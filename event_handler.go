package main

import "time"
import "github.com/godbus/dbus"
import "github.com/desertbit/timer"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/mousebind"
import "github.com/BurntSushi/xgbutil/keybind"

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

type Timers struct {
	Uid          uint
	Notification *Notification
	TimeoutTimer *timer.Timer
	PopupTimer   *timer.Timer
	StopNow      chan struct{}
}

func NewTimers(uid uint, n *Notification, d time.Duration, p time.Duration) *Timers {
	t := &Timers{
		Uid:          uid,
		Notification: n,
		StopNow:      make(chan struct{}),
	}
	if d < 0 {
		t.TimeoutTimer = timer.NewStoppedTimer()
	} else {
		t.TimeoutTimer = timer.NewTimer(d)
	}
	if p < 0 {
		t.PopupTimer = timer.NewStoppedTimer()
	} else {
		t.PopupTimer = timer.NewTimer(p)
	}
	return t
}

func (t *Timers) Stop() {
	t.PopupTimer.Stop()
	t.TimeoutTimer.Stop()
	t.StopNow <- struct{}{}
}

func (t *Timers) Reset(d time.Duration, p time.Duration) {
	t.PopupTimer.Reset(p)
	if d < 0 {
		t.TimeoutTimer = timer.NewStoppedTimer()
	} else {
		t.TimeoutTimer.Reset(d)
	}
}

type EventHandler struct {
	conn             *dbus.Conn
	nextUid          uint
	closeChan        chan uint32 // Used for HandleClose
	closedChan       chan bool   // Used for HandleClose
	actionChan       chan Action
	expiryChan       chan Expiry
	notificationChan chan *Notification
	timers           map[uint32]*Timers
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
		expiryChan:       make(chan Expiry),
		actionChan:       make(chan Action),
		notificationChan: make(chan *Notification),
		timers:           map[uint32]*Timers{},
		display: &PopupDisplay{
			x:      X,
			active: map[uint]*Popup{},
		},
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
	t := e.timers[id]
	delete(e.timers, id)
	go t.Stop()
	e.display.Destroy(uid)
	e.EmitClosed(id, reason)
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
			t := e.timers[n.Id]
			if t != nil {
				// We have seen this before, so stop the current timers
				go t.Stop()
				old = t.Uid
			}
			e.display.Show(old, uid, n, e.GetContextMenuFunc(), e.GetCloseFunc())
			t = NewTimers(uid, n, maxTimeout, maxAge)
			go func(uid uint) {
				for {
					select {
					case <-t.TimeoutTimer.C:
						e.expiryChan <- Expiry{ExpiryTimeout, n.Id, uid}
					case <-t.PopupTimer.C:
						e.expiryChan <- Expiry{ExpiryPopup, n.Id, uid}
					case <-t.StopNow:
						break
					}
				}
			}(t.Uid)
			e.timers[n.Id] = t
			e.display.Draw()
		case exp := <-e.expiryChan:
			t := e.timers[exp.Id]
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
		case a := <-e.actionChan:
			switch a.Type {
			case ActionShowAll:
				for _, t := range e.timers {
					e.display.Show(t.Uid, t.Uid, t.Notification, e.GetContextMenuFunc(), e.GetCloseFunc())
					if t.Notification.Hints.Urgency != UrgencyCritical {
						t.PopupTimer.Reset(conf.Core.MaxPopupAge.Duration)
					}
				}
				e.display.Draw()
			case ActionCloseLatest:
				target := uint32(0)
				uid := uint(0)
				for id, t := range e.timers {
					if t.Uid > uid {
						uid = t.Uid
						target = id
					}
				}
				if target != 0 {
					e.deleteNotification(target, uid, 2)
				}
			case ActionCloseOne:
				if t := e.timers[a.Target]; t != nil {
					e.deleteNotification(a.Target, t.Uid, 2)
				}
			case ActionContextMenu:
				if t := e.timers[a.Target]; t != nil {
					go execMixedSelector(t.Notification.Actions, t.Notification.Body.Hyperlinks, func(action_key string) {
						e.EmitAction(t.Notification.Id, action_key)
						if !t.Notification.Hints.Resident {
							e.expiryChan <- Expiry{ExpiryAction, a.Target, t.Uid}
						}
					})
				}
			}
		case id := <-e.closeChan:
			if t, ok := e.timers[id]; ok {
				e.deleteNotification(id, t.Uid, 3)
				e.closedChan <- true
			} else {
				e.closedChan <- false
			}
		}
	}
}
