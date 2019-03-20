package main

import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xgraphics"
import "github.com/BurntSushi/xgbutil/xwindow"
import "github.com/BurntSushi/xgbutil/ewmh"
import "github.com/BurntSushi/xgbutil/mousebind"

var (
	fontBold    = mustReadFont("/usr/share/fonts/truetype/dejavu/DejaVuSansMono-Bold.ttf")
	fontRegular = mustReadFont("/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf")
)

//var bg = xgraphics.BGRA{B: 0x55, G: 0x55, R: 0x00, A: 0xff}
//var bgUrgent = xgraphics.BGRA{B: 0x00, G: 0x11, R: 0x66, A: 0xff}

type TextLine struct {
	Text   string
	Height int
}

func ximgFromNotification(X *xgbutil.XUtil, n *Notification, body string) *xgraphics.Image {
	key := "low"
	switch n.Hints.Urgency {
	case UrgencyCritical:
		key = "critical"
	case UrgencyNormal:
		key = "normal"
	case UrgencyLow:
		key = "low"
	}
	style := conf.Styles[key]
	fontSize := float64(conf.Core.FontSize)
	padding := *style.Padding
	bgColor := style.Background.BGRA
	fgColor := style.Foreground.BGRA
	notificationWidth := conf.Core.Width

	fontWidthOracle := func(s string) int {
		w, _ := xgraphics.Extents(fontRegular, fontSize, s)
		return w
	}

	summary := maxWidth(n.AppName+": "+n.Summary, notificationWidth, func(s string) int {
		w, _ := xgraphics.Extents(fontBold, fontSize, s)
		return w
	})
	_, firsth := xgraphics.Extents(fontBold, fontSize, summary)

	chunks := []TextLine{}
	height := firsth

	for {
		text := maxWidth(body, notificationWidth, fontWidthOracle)
		_, h := xgraphics.Extents(fontRegular, fontSize, text)
		chunks = append(chunks, TextLine{text, h})
		height += h
		body = body[len(text):]
		if len(body) == 0 {
			break
		}
	}
	// create canvas
	ximg := ximgWithProps(X, padding, height, notificationWidth, 2, bgColor, fgColor)
	h := 0
	// draw text
	_, _, _ = ximg.Text(padding, padding, fgColor, fontSize, fontBold, summary)
	for _, line := range chunks {
		_, _, _ = ximg.Text(padding, padding+firsth+h, fgColor, fontSize, fontRegular, line.Text)
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
	p := &Popup{}
	p.x = x
	p.order = order
	p.Update(n)
	return p
}

func (p *Popup) Shown() bool {
	return p.window != nil
}

func (p *Popup) Height() int {
	g, _ := p.window.Geometry()
	return g.Height()
}

func (p *Popup) Update(n *Notification) {
	p.notification = n
	body, links := TextInfoFromString(n.Body)
	ximg := ximgFromNotification(p.x, n, body)
	p.links = links
	p.window = ximg.XShow()
	// care: this should be done before drawing anything because otherwise
	// we would get some glitch
	ewmh.WmWindowTypeSet(p.x, p.window.Id, []string{"_NET_WM_WINDOW_TYPE_NOTIFICATION"})
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
