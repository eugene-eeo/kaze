package main

import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/mousebind"

type PopupDisplay struct {
	x      *xgbutil.XUtil
	active map[UID]*Popup
}

func NewPopupDisplay(x *xgbutil.XUtil) *PopupDisplay {
	return &PopupDisplay{
		x:      x,
		active: map[UID]*Popup{},
	}
}

func bindMouseCallbacks(X *xgbutil.XUtil, popup *Popup, ctxMenuFunc func(*Notification), closeFunc func(*Notification)) {
	cmenu := mousebind.ButtonPressFun(func(X *xgbutil.XUtil, e xevent.ButtonPressEvent) { ctxMenuFunc(popup.notification) })
	close := mousebind.ButtonPressFun(func(X *xgbutil.XUtil, e xevent.ButtonPressEvent) { closeFunc(popup.notification) })
	cmenu.Connect(X, popup.window.Id, conf.Bindings.Filter, false, true)
	close.Connect(X, popup.window.Id, conf.Bindings.CloseOne, false, true)
}

func (p *PopupDisplay) Show(old UID, uid UID, n *Notification, ctxMenuFunc func(*Notification), closeFunc func(*Notification)) {
	popup := p.active[old]
	if popup == nil {
		// not seen before or currently invisible
		popup = NewPopup(p.x, uint(uid), n)
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
	hide := false
	for _, pair := range order {
		if popup := p.active[pair.Uid]; popup != nil {
			if !hide {
				// If we are in showing mode
				popup.Move(conf.Style.XOffset, height)
				height += popup.Height() - conf.Style.BorderWidth
				if height >= conf.Core.MaxHeight {
					hide = true
				}
			} else {
				// otherwise we need to close the window
				p.Close(pair.Uid)
			}
		}
	}
}

func (p *PopupDisplay) Close(uid UID) {
	p.Destroy(uid)
}

func (p *PopupDisplay) Destroy(uid UID) {
	popup := p.active[uid]
	if popup != nil {
		popup.Close()
		delete(p.active, uid)
	}
}

func (p *PopupDisplay) FirstVisible(order []*UidPair) *UidPair {
	for _, pair := range order {
		if p.active[pair.Uid] != nil {
			return pair
		}
	}
	return nil
}
