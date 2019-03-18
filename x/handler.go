package x

import "os"
import "image"
import "github.com/godbus/dbus"
import "github.com/eugene-eeo/kaze/libkaze"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xwindow"
import "github.com/BurntSushi/xgbutil/xgraphics"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/ewmh"

const notificationWidth = 200
const notificationHeight = 50
const fontPath = "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"
const fontSize = 14

var transparent = xgraphics.BGRA{B: 0xff, G: 0x00, R: 0x00, A: 0x00}
var bg = xgraphics.BGRA{B: 0xff, G: 0x66, R: 0x33, A: 0xff}
var fg = xgraphics.BGRA{B: 0xff, G: 0xff, R: 0xff, A: 0xff}

var monitorWidth = 1920
var monitorHeight = 1080

type XHandler struct {
	X       *xgbutil.XUtil
	windows map[uint32]*xwindow.Window
}

func NewXHandler() *XHandler {
	X, err := xgbutil.NewConn()
	if err != nil {
		panic(err)
	}
	go xevent.Main(X)
	return &XHandler{
		X:       X,
		windows: map[uint32]*xwindow.Window{},
	}
}

func (_ *XHandler) Capabilities() []string {
	return []string{"body"}
}

func (h *XHandler) HandleNotification(n *libkaze.Notification) {
	fontReader, err := os.Open(fontPath)
	if err != nil {
		panic(err)
	}
	// parse font
	font, err := xgraphics.ParseFont(fontReader)
	font = xgraphics.MustFont(font, err)

	// create canvas
	ximg := xgraphics.New(h.X, image.Rect(0, 0, notificationWidth, notificationHeight))
	ximg.For(func(x, y int) xgraphics.BGRA {
		return bg
	})

	_, _, err = ximg.Text(10, 10, fg, fontSize, font, n.Summary)
	if err != nil {
		panic(err)
	}

	_, firsth := xgraphics.Extents(font, fontSize, n.Summary)
	_, _, err = ximg.Text(10, 10+firsth, fg, fontSize, font, n.Body)
	if err != nil {
		panic(err)
	}
	secw, sech := xgraphics.Extents(font, fontSize, n.Body)
	bounds := image.Rect(10, 10+firsth, 10+secw, 10+firsth+sech)

	win := h.windows[n.Id]
	if win == nil {
		// if we cannot find a window
		win = ximg.XShow()
		ewmh.WmWindowTypeSet(h.X, win.Id, []string{"_NET_WM_WINDOW_TYPE_NOTIFICATION"})
		win.Move(monitorWidth-notificationWidth, 20)
		h.windows[n.Id] = win
	}
	ximg.SubImage(bounds).(*xgraphics.Image).XDraw()
	ximg.XPaint(win.Id)
}

func (h *XHandler) HandleClose(id uint32) *dbus.Error {
	if h.windows[id] != nil {
		h.windows[id].Destroy()
		delete(h.windows, id)
	}
	return nil
}

func (h *XHandler) HandleTimeout(id uint32) {
	if h.windows[id] != nil {
		h.windows[id].Destroy()
		delete(h.windows, id)
	}
}
