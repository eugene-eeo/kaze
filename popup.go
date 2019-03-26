package main

import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xgraphics"
import "github.com/BurntSushi/xgbutil/xwindow"
import "github.com/BurntSushi/xgbutil/ewmh"
import "github.com/eugene-eeo/kaze/render"

type TextLine struct {
	Text   string
	Height int
}

func ximgFromNotification(X *xgbutil.XUtil, n *Notification) *xgraphics.Image {
	bg := conf.Style.NormalBg.BGRA
	switch n.Hints.Urgency {
	case UrgencyCritical:
		bg = conf.Style.CriticalBg.BGRA
	case UrgencyLow:
		bg = conf.Style.LowBg.BGRA
	}
	fg := conf.Style.Fg.BGRA
	fontSize := float64(conf.Style.FontSize)
	padding := conf.Style.Padding
	notificationWidth := conf.Style.Width

	flBold := render.FontList{fontBold, fontFallback}
	flRegular := render.FontList{fontRegular, fontFallback}

	summary := render.TextBoxWithMaxWidth(n.AppName+": "+n.Summary, flBold, fontSize, notificationWidth)[0]
	textbox := render.TextBoxWithMaxWidth(n.Body.Text, flRegular, fontSize, notificationWidth)

	height := summary.Height
	for _, line := range textbox {
		height += line.Height
	}

	// create canvas
	ximg := ximgWithProps(X, padding,
		height, notificationWidth,
		conf.Style.BorderWidth, bg,
		conf.Style.BorderColor.BGRA)
	// draw text
	h := padding
	x := padding
	for _, c := range summary.Chunks {
		x, _, _ = ximg.Text(x, h, fg, fontSize, c.Font, string(c.Rune))
	}
	h += summary.Height
	for _, line := range textbox {
		x := padding
		for _, c := range line.Chunks {
			x, _, _ = ximg.Text(x, h, fg, fontSize, c.Font, string(c.Rune))
		}
		h += line.Height
	}
	return ximg
}

type Popup struct {
	order        uint
	window       *xwindow.Window
	notification *Notification
	x            *xgbutil.XUtil
	links        []Hyperlink
}

func NewPopup(x *xgbutil.XUtil, order uint, n *Notification) *Popup {
	p := &Popup{x: x, order: order}
	ximg := ximgFromNotification(p.x, n)
	p.notification = n
	p.window = ximg.XShow()
	// care: this should be done before drawing anything because otherwise
	// we would get some glitch
	ewmh.WmWindowTypeSet(p.x, p.window.Id, []string{"_NET_WM_WINDOW_TYPE_NOTIFICATION"})
	ximg.XDraw()
	ximg.XPaint(p.window.Id)
	return p
}

func (p *Popup) Height() int {
	g, _ := p.window.Geometry()
	return g.Height()
}

func (p *Popup) Update(n *Notification) {
	if p.notification != n {
		ximg := ximgFromNotification(p.x, n)
		p.window.Resize(ximg.Bounds().Max.X, ximg.Bounds().Max.Y)
		ximg.XDraw()
		ximg.XPaint(p.window.Id)
		p.notification = n
	}
}

func (p *Popup) Move(x, y int) {
	p.window.Move(x, y)
}

func (p *Popup) Close() {
	p.window.Detach()
	p.window.Destroy()
}
