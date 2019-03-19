package main

import "time"
import "sort"

import "github.com/godbus/dbus"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/mousebind"

const popupMaxAge = 3500 * time.Millisecond

type orderIdPair struct {
	id    uint32
	order uint
}

type XHandler struct {
	X               *xgbutil.XUtil
	popups          map[uint32]*Popup
	Wrapper         *HandlerWrapper
	uid             uint
	removeChan      chan uint32
	popupRemoveChan chan orderIdPair
}

func NewXHandler() *XHandler {
	X, err := xgbutil.NewConn()
	if err != nil {
		panic(err)
	}
	mousebind.Initialize(X)
	go xevent.Main(X)
	return &XHandler{
		X:               X,
		popups:          map[uint32]*Popup{},
		removeChan:      make(chan uint32),
		popupRemoveChan: make(chan orderIdPair),
	}
}

func (_ *XHandler) Capabilities() []string {
	return []string{"body", "actions"}
}

func (h *XHandler) HandleNotification(n *Notification) {
	popup := h.popups[n.Id]
	if popup == nil {
		// If we cannot find one then we need to increment uid
		h.uid++
		popup = NewPopup(h.X, h.uid, n)
		h.popups[n.Id] = popup
		cb := mousebind.ButtonPressFun(func(X *xgbutil.XUtil, e xevent.ButtonPressEvent) {
			h.Wrapper.SilentNotificationClose(n.Id)
			h.HandleClose(n.Id)
		})
		cb.Connect(h.X, popup.window.Id, "1", false, true)
		uid := h.uid
		go func() {
			time.Sleep(popupMaxAge)
			h.popupRemoveChan <- orderIdPair{n.Id, uid}
		}()
	} else {
		popup.Update(n)
	}
	h.repaint()
}

func (h *XHandler) repaint() {
	ids := make([]uint32, 0, len(h.popups))
	for id, popup := range h.popups {
		if popup.Shown() {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		return h.popups[ids[i]].order < h.popups[ids[j]].order
	})
	height := 20 + padding
	x := monitorWidth - (notificationWidth + 2*padding + 2*2) - 10
	for _, id := range ids {
		popup := h.popups[id]
		popup.Move(x, height)
		height += popup.height
		height += padding
	}
}

func (h *XHandler) Loop() {
	for {
		select {
		case id := <-h.removeChan:
			if h.popups[id] != nil {
				h.popups[id].Close()
				delete(h.popups, id)
				h.repaint()
			}
		// this is for popup timeouts
		// dont delete the entry, but just repaint and close window
		case pair := <-h.popupRemoveChan:
			if h.popups[pair.id] != nil {
				popup := h.popups[pair.id]
				if popup.order == pair.order {
					popup.Close()
					h.repaint()
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
