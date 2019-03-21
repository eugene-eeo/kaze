package main

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
	cmenu.Connect(X, popup.window.Id, conf.Bindings.Filter, false, true)
	close.Connect(X, popup.window.Id, conf.Bindings.CloseOne, false, true)
}

func (p *PopupDisplay) Show(old uint, uid uint, n *Notification, ctxMenuFunc func(*Notification), closeFunc func(*Notification)) {
	popup := p.active[old]
	if popup == nil {
		// not seen before or currently invisible
		popup = NewPopup(p.x, uid, n)
		bindMouseCallbacks(p.x, popup, ctxMenuFunc, closeFunc)
	} else {
		// otherwise it is currently visible
		popup.Update(n)
	}
	delete(p.active, old)
	p.active[uid] = popup
}

func (p *PopupDisplay) Draw(order []*UidPair) {
	height := conf.Style.YOffset
	for _, pair := range order {
		if popup := p.active[pair.Uid]; popup != nil {
			popup.Move(conf.Style.XOffset, height)
			height += popup.Height() - conf.Style.BorderWidth
		}
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
