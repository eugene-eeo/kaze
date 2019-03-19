package main

import "fmt"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xgraphics"
import "github.com/BurntSushi/xgbutil/xwindow"
import "github.com/BurntSushi/xgbutil/ewmh"
import "github.com/BurntSushi/xgbutil/mousebind"

// TODO: use config file for these
const notificationWidth = 300
const fontSize = 14
const padding = 10
const monitorWidth = 1920
const monitorHeight = 1080

var font = mustReadFont("/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf")
var fg = xgraphics.BGRA{B: 0xff, G: 0xff, R: 0xff, A: 0xff}
var bg = xgraphics.BGRA{B: 0x55, G: 0x55, R: 0x00, A: 0xff}
var bgUrgent = xgraphics.BGRA{B: 0x00, G: 0x11, R: 0x66, A: 0xff}

func ximgFromNotification(X *xgbutil.XUtil, n *Notification) *xgraphics.Image {
	fontWidthOracle := func(s string) int {
		w, _ := xgraphics.Extents(font, fontSize, s)
		return w
	}

	summary := maxWidth(fmt.Sprintf("%s: %s", n.AppName, n.Summary), notificationWidth, fontWidthOracle)
	bodyText := maxWidth(n.Body, notificationWidth, fontWidthOracle)

	_, firsth := xgraphics.Extents(font, fontSize, summary)
	_, sech := xgraphics.Extents(font, fontSize, bodyText)

	// create canvas
	bgColor := bg
	if n.Hints.Urgency == UrgencyCritical {
		bgColor = bgUrgent
	}
	ximg := ximgWithProps(X, padding, firsth+sech, notificationWidth, 2, bgColor, fg)
	// draw text
	_, _, _ = ximg.Text(padding, padding, fg, fontSize, font, summary)
	_, _, _ = ximg.Text(padding, padding+firsth, fg, fontSize, font, bodyText)
	return ximg
}

type Popup struct {
	order        uint
	height       int
	window       *xwindow.Window
	notification *Notification
	x            *xgbutil.XUtil
}

func NewPopup(x *xgbutil.XUtil, order uint, n *Notification) *Popup {
	p := &Popup{}
	p.x = x
	p.order = order
	p.Update(n)
	ewmh.WmWindowTypeSet(x, p.window.Id, []string{"_NET_WM_WINDOW_TYPE_NOTIFICATION"})
	return p
}

func (p *Popup) Shown() bool {
	return p.window != nil
}

func (p *Popup) Update(n *Notification) {
	p.notification = n
	ximg := ximgFromNotification(p.x, n)
	p.window = ximg.XShow()
	p.height = ximg.Rect.Max.Y
	ximg.XDraw()
	ximg.XPaint(p.window.Id)
}

func (p *Popup) Move(x, y int) {
	if p.window != nil {
		p.window.Move(x, y)
	}
}

func (p *Popup) Close() {
	if p.window != nil {
		mousebind.Detach(p.x, p.window.Id)
		p.window.Destroy()
		p.window = nil
	}
}
