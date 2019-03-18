package main

import "time"
import "sort"

import "github.com/godbus/dbus"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/mousebind"

const popupMaxAge = 3500 * time.Millisecond

type XHandler struct {
	X       *xgbutil.XUtil
	entries map[uint32]*Popup
	Wrapper *HandlerWrapper
	uid     uint
}

func NewXHandler() *XHandler {
	X, err := xgbutil.NewConn()
	if err != nil {
		panic(err)
	}
	mousebind.Initialize(X)
	go xevent.Main(X)
	return &XHandler{
		X:       X,
		entries: map[uint32]*Popup{},
	}
}

func (_ *XHandler) Capabilities() []string {
	return []string{"body", "actions"}
}

func (h *XHandler) HandleNotification(n *Notification) {
	winOrder := h.entries[n.Id]
	if winOrder == nil {
		// If we cannot find one then we need to increment uid
		h.uid++
		popup := NewPopup(h.X, h.uid, n)
		h.entries[n.Id] = popup
		cb := mousebind.ButtonPressFun(func(X *xgbutil.XUtil, e xevent.ButtonPressEvent) {
			popup.window.Destroy()
			delete(h.entries, n.Id)
			h.Wrapper.SilentNotificationClose(n.Id)
			h.repaint()
		})
		cb.Connect(h.X, popup.window.Id, "1", false, true)
	} else {
		h.entries[n.Id].Update(n)
	}
	h.repaint()
}

func (h *XHandler) close(id uint32) {
	h.Wrapper.SilentNotificationClose(id)
}

func (h *XHandler) repaint() {
	ids := make([]uint32, len(h.entries))
	i := 0
	for id, _ := range h.entries {
		ids[i] = id
		i++
	}
	sort.Slice(ids, func(i, j int) bool {
		return h.entries[ids[i]].order < h.entries[ids[j]].order
	})
	height := 20 + padding
	x := monitorWidth - (notificationWidth + 2*padding + 2*2) - 10
	for _, id := range ids {
		popup := h.entries[id]
		popup.Move(x, height)
		height += popup.height
		height += padding
	}
}

func (h *XHandler) HandleClose(id uint32) *dbus.Error {
	if h.entries[id] != nil {
		h.entries[id].window.Destroy()
		delete(h.entries, id)
	}
	h.repaint()
	return nil
}

func (h *XHandler) HandleTimeout(id uint32) {
	if h.entries[id] != nil {
		h.entries[id].window.Destroy()
		delete(h.entries, id)
	}
	h.repaint()
}
