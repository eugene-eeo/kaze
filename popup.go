package main

import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xgraphics"
import "github.com/BurntSushi/xgbutil/xwindow"
import "github.com/BurntSushi/xgbutil/ewmh"

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

	summary := maxWidth(n.AppName+": "+n.Summary, notificationWidth, fontBold, fontSize)
	_, firsth := xgraphics.Extents(fontBold, fontSize, summary)

	chunks := []TextLine{}
	height := firsth
	body := n.Body.Text

	for {
		text := maxWidth(body, notificationWidth, fontRegular, fontSize)
		_, h := xgraphics.Extents(fontRegular, fontSize, text)
		chunks = append(chunks, TextLine{text, h})
		height += h
		body = body[len(text):]
		if len(body) == 0 {
			break
		}
	}
	// create canvas
	ximg := ximgWithProps(X, padding,
		height, notificationWidth,
		conf.Style.BorderWidth, bg,
		conf.Style.BorderColor.BGRA)
	h := 0
	// draw text
	_, _, _ = ximg.Text(padding, padding, fg, fontSize, fontBold, summary)
	for _, line := range chunks {
		_, _, _ = ximg.Text(padding, padding+firsth+h, fg, fontSize, fontRegular, line.Text)
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
