package x

import "time"
import "fmt"
import "os"
import "sort"
import "image"
import "github.com/godbus/dbus"
import "github.com/eugene-eeo/kaze/libkaze"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xwindow"
import "github.com/BurntSushi/xgbutil/xgraphics"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/ewmh"
import "github.com/BurntSushi/xgbutil/mousebind"

const popupMaxAge = 3500 * time.Millisecond
const notificationWidth = 300
const fontPath = "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"
const fontSize = 14

var bg = xgraphics.BGRA{B: 0x55, G: 0x55, R: 0x00, A: 0xff}
var fg = xgraphics.BGRA{B: 0xff, G: 0xff, R: 0xff, A: 0xff}

var padding = 10
var monitorWidth = 1920
var monitorHeight = 1080

type windowOrder struct {
	order  uint
	window *xwindow.Window
	height int
}

type XHandler struct {
	X       *xgbutil.XUtil
	windows map[uint32]*windowOrder
	Wrapper *libkaze.HandlerWrapper
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
		windows: map[uint32]*windowOrder{},
	}
}

func (_ *XHandler) Capabilities() []string {
	return []string{"body", "actions"}
}

func (h *XHandler) HandleNotification(n *libkaze.Notification) {
	fontReader, err := os.Open(fontPath)
	if err != nil {
		panic(err)
	}
	// parse font
	font, err := xgraphics.ParseFont(fontReader)
	font = xgraphics.MustFont(font, err)

	summary := maxWidth(fmt.Sprintf("%s: %s", n.AppName, n.Summary), notificationWidth, func(s string) int {
		w, _ := xgraphics.Extents(font, fontSize, s)
		return w
	})

	bodyText := maxWidth(n.Body, notificationWidth, func(s string) int {
		w, _ := xgraphics.Extents(font, fontSize, s)
		return w
	})

	firstw, firsth := xgraphics.Extents(font, fontSize, summary)
	secw, sech := xgraphics.Extents(font, fontSize, bodyText)

	// create canvas
	ximg := ximgWithProps(h.X, padding, firsth+sech, notificationWidth, 2, bg, fg)

	_, _, err = ximg.Text(padding, padding, fg, fontSize, font, summary)
	if err != nil {
		panic(err)
	}

	_, _, err = ximg.Text(padding, padding+firsth, fg, fontSize, font, bodyText)
	if err != nil {
		panic(err)
	}

	var win *xwindow.Window
	winOrder := h.windows[n.Id]
	if winOrder == nil {
		// if we cannot find a window
		win = ximg.XShow()
		ewmh.WmWindowTypeSet(h.X, win.Id, []string{"_NET_WM_WINDOW_TYPE_NOTIFICATION"})
		h.uid++
		id := n.Id
		uid := h.uid
		h.windows[n.Id] = &windowOrder{h.uid, win, 2*padding + firsth + sech}
		// automatically close and destroy window, but do not emit the close
		// notification action
		go func() {
			time.Sleep(popupMaxAge)
			w := h.windows[id]
			if w != nil && w.order == uid {
				w.window.Destroy()
				delete(h.windows, id)
				h.repaint()
			}
		}()
		cb := mousebind.ButtonPressFun(func(x *xgbutil.XUtil, e xevent.ButtonPressEvent) {
			win.Destroy()
			delete(h.windows, id)
			h.close(id)
			h.repaint()
		})
		cb.Connect(h.X, win.Id, "1", false, true)
	} else {
		win = winOrder.window
	}
	bounds := image.Rect(10, 10+firsth, 10+max(firstw, secw), 10+firsth+sech)
	ximg.SubImage(bounds).(*xgraphics.Image).XDraw()
	ximg.XPaint(win.Id)
	h.repaint()
}

func (h *XHandler) close(id uint32) {
	h.Wrapper.SilentNotificationClose(id)
}

func (h *XHandler) repaint() {
	ids := make([]uint32, len(h.windows))
	i := 0
	for id, _ := range h.windows {
		ids[i] = id
		i++
	}
	sort.Slice(ids, func(i, j int) bool {
		return h.windows[ids[i]].order < h.windows[ids[j]].order
	})
	height := 20 + 10
	x := monitorWidth - (notificationWidth + 2*padding + 2*2) - 10
	for _, id := range ids {
		windowInfo := h.windows[id]
		windowInfo.window.Move(x, height)
		height += windowInfo.height
		height += padding
	}
}

func (h *XHandler) HandleClose(id uint32) *dbus.Error {
	if h.windows[id] != nil {
		h.windows[id].window.Destroy()
		delete(h.windows, id)
	}
	h.repaint()
	return nil
}

func (h *XHandler) HandleTimeout(id uint32) {
	if h.windows[id] != nil {
		h.windows[id].window.Destroy()
		delete(h.windows, id)
	}
	h.repaint()
}
