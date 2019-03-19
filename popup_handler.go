package main

import "time"
import "sort"

import "github.com/godbus/dbus"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/mousebind"
import "github.com/BurntSushi/xgbutil/keybind"
import "github.com/desertbit/timer"

const popupMaxAge = 3500 * time.Millisecond
const (
	ActionShowAll = iota
	ActionCloseLatest
	ActionRepaint
)

type actionIdPair struct {
	id  uint32
	key string
}

type orderIdPair struct {
	id    uint32
	order uint
}

type XHandler struct {
	X                 *xgbutil.XUtil
	popups            map[uint32]*Popup
	Wrapper           *HandlerWrapper
	uid               uint
	removeChan        chan uint32
	popupRemoveChan   chan orderIdPair
	actionChan        chan int
	actionInvokedChan chan actionIdPair
	closeShowAllTimer *timer.Timer
}

func NewXHandler() *XHandler {
	X, err := xgbutil.NewConn()
	if err != nil {
		panic(err)
	}
	mousebind.Initialize(X)
	keybind.Initialize(X)
	handler := &XHandler{
		X:                 X,
		popups:            map[uint32]*Popup{},
		removeChan:        make(chan uint32),
		popupRemoveChan:   make(chan orderIdPair),
		actionChan:        make(chan int),
		actionInvokedChan: make(chan actionIdPair),
		closeShowAllTimer: timer.NewTimer(0),
	}

	showAllCb := keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		handler.actionChan <- ActionShowAll
	})

	closeLatestCb := keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		handler.actionChan <- ActionCloseLatest
	})

	showAllCb.Connect(X, X.RootWin(), "Mod1-Space", true)
	closeLatestCb.Connect(X, X.RootWin(), "Mod1-Shift-Space", true)

	go xevent.Main(X)
	return handler
}

func (_ *XHandler) Capabilities() []string {
	return []string{
		"body",
		"actions",
		"persistence",
		"action-icons",
		"body-hyperlinks",
		"body-images",
		"body-markup",
		"icon-multi",
		"icon-static",
		"sound",
	}
}

func (h *XHandler) bindMousekeys(p *Popup) {
	// close event
	closeWindow := mousebind.ButtonPressFun(func(X *xgbutil.XUtil, e xevent.ButtonPressEvent) {
		h.Wrapper.SilentNotificationClose(p.notification.Id)
		h.HandleClose(p.notification.Id)
	})
	// actions
	actions := mousebind.ButtonPressFun(func(X *xgbutil.XUtil, e xevent.ButtonPressEvent) {
		go execMixedSelector(p.notification.Actions, p.links, func(action_key string) {
			h.actionInvokedChan <- actionIdPair{p.notification.Id, action_key}
		})
	})
	closeWindow.Connect(h.X, p.window.Id, "3", false, true)
	actions.Connect(h.X, p.window.Id, "1", false, true)
}

func (h *XHandler) HandleNotification(n *Notification) {
	popup := h.popups[n.Id]
	if popup == nil {
		// If we cannot find one then we need to increment uid
		h.uid++
		popup = NewPopup(h.X, h.uid, n)
		h.popups[n.Id] = popup
		h.bindMousekeys(popup)
		uid := h.uid
		go func() {
			time.Sleep(popupMaxAge)
			h.popupRemoveChan <- orderIdPair{n.Id, uid}
		}()
	} else {
		popup.Update(n)
	}
	h.actionChan <- ActionRepaint
}

func (h *XHandler) repaint() {
	ids := make([]uint32, 0, len(h.popups))
	for id, popup := range h.popups {
		if popup.Shown() {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		return h.popups[ids[i]].order > h.popups[ids[j]].order
	})
	height := 20 + padding
	x := monitorWidth - (notificationWidth + 2*padding + 2*2) - 10
	for _, id := range ids {
		popup := h.popups[id]
		popup.Move(x, height)
		height += popup.height - 2
	}
}

func (h *XHandler) Loop() {
	for {
		select {
		case <-h.closeShowAllTimer.C:
			for _, popup := range h.popups {
				// close all non-critical notifications
				if popup.notification.Hints.Urgency != UrgencyCritical && popup.Shown() {
					id := popup.notification.Id
					order := popup.order
					go func() {
						h.popupRemoveChan <- orderIdPair{id, order}
					}()
				}
			}
		case id := <-h.removeChan:
			popup := h.popups[id]
			if popup != nil {
				popup.Close()
				delete(h.popups, id)
				h.repaint()
			}
		// this is for popup timeouts
		// dont delete the entry, but just repaint and close window
		case pair := <-h.popupRemoveChan:
			popup := h.popups[pair.id]
			if popup != nil && popup.order == pair.order {
				popup.Close()
				h.repaint()
			}
		case action := <-h.actionChan:
			switch action {
			case ActionRepaint:
				h.repaint()
			case ActionShowAll:
				h.closeShowAllTimer.Reset(popupMaxAge)
				for _, popup := range h.popups {
					if !popup.Shown() {
						popup.Update(popup.notification)
						h.bindMousekeys(popup)
					}
				}
				h.repaint()
			case ActionCloseLatest:
				maxId := uint32(0)
				for id, _ := range h.popups {
					if id > maxId {
						maxId = id
					}
				}
				popup := h.popups[maxId]
				if popup != nil {
					go func() {
						h.Wrapper.SilentNotificationClose(popup.notification.Id)
						h.HandleClose(popup.notification.Id)
					}()
				}
			}
		case pair := <-h.actionInvokedChan:
			if popup, ok := h.popups[pair.id]; ok {
				h.Wrapper.ActionInvoked(pair.id, pair.key)
				if !popup.notification.Hints.Resident {
					popup.Close()
					delete(h.popups, pair.id)
				}
			}
		}
	}
}

func (h *XHandler) HandleClose(id uint32) *dbus.Error {
	h.HandleTimeout(id)
	return nil
}

func (h *XHandler) HandleTimeout(id uint32) {
	h.removeChan <- id
}
