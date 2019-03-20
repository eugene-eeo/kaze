package main

import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xgraphics"
import "github.com/BurntSushi/xgbutil/xwindow"
import "github.com/BurntSushi/xgbutil/ewmh"

var (
	fontBold    = mustReadFont("/usr/share/fonts/truetype/dejavu/DejaVuSansMono-Bold.ttf")
	fontRegular = mustReadFont("/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf")
)

type TextLine struct {
	Text   string
	Height int
}

func ximgFromNotification(X *xgbutil.XUtil, n *Notification) *xgraphics.Image {
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
	body := n.Body.Text

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
	ximg := ximgFromNotification(p.x, n)
	ximg.Window(p.window.Id)
	ximg.XDraw()
	ximg.XPaint(p.window.Id)
}

func (p *Popup) Move(x, y int) {
	p.window.Move(x, y)
}

func (p *Popup) Close() {
	p.window.Detach()
	p.window.Destroy()
}
