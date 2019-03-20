package main

import "sort"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/mousebind"

type PopupDisplay struct {
	x      *xgbutil.XUtil
	active map[uint]*Popup
}

func bindMouseCallbacks(X *xgbutil.XUtil, popup *Popup, ctxMenuFunc func(*Notification), closeFunc func(*Notification)) {
	cmenu := mousebind.ButtonPressFun(func(X *xgbutil.XUtil, e xevent.ButtonPressEvent) { go ctxMenuFunc(popup.notification) })
	close := mousebind.ButtonPressFun(func(X *xgbutil.XUtil, e xevent.ButtonPressEvent) { go closeFunc(popup.notification) })
	cmenu.Connect(X, popup.window.Id, "1", false, true)
	close.Connect(X, popup.window.Id, "3", false, true)
}

func (p *PopupDisplay) Show(old uint, uid uint, n *Notification, ctxMenuFunc func(*Notification), closeFunc func(*Notification)) {
	popup := p.active[old]
	if popup == nil {
		// not seen before
		popup = NewPopup(p.x, uid, n)
		bindMouseCallbacks(p.x, popup, ctxMenuFunc, closeFunc)
	} else {
		// otherwise it is currently visible
		popup.Update(n)
	}
	delete(p.active, old)
	p.active[uid] = popup
}

func (p *PopupDisplay) Draw() {
	ids := make([]uint, 0, len(p.active))
	for id, popup := range p.active {
		if popup != nil {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		a := p.active[ids[i]]
		b := p.active[ids[j]]
		return a.notification.Hints.Urgency > b.notification.Hints.Urgency || a.order > b.order
	})
	height := conf.Core.YOffset
	for _, id := range ids {
		popup := p.active[id]
		popup.Move(conf.Core.XOffset, height)
		height += popup.Height() - 2
	}
}

func (p *PopupDisplay) Close(uid uint) {
	if popup := p.active[uid]; popup != nil {
		popup.Close()
		p.active[uid] = nil
	}
}

func (p *PopupDisplay) Destroy(uid uint) {
	popup := p.active[uid]
	if popup != nil {
		popup.Close()
		delete(p.active, uid)
	}
}