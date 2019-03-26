package main

import "github.com/godbus/dbus"
import "github.com/eugene-eeo/kaze/config"
import "github.com/BurntSushi/freetype-go/freetype/truetype"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/keybind"
import "github.com/BurntSushi/xgbutil/mousebind"

var conf *config.Config
var fontBold *truetype.Font
var fontRegular *truetype.Font
var fontFallback *truetype.Font

func newServer(conn *dbus.Conn) *Server {
	X, err := xgbutil.NewConn()
	if err != nil {
		panic(err)
	}
	keybind.Initialize(X)
	mousebind.Initialize(X)

	server := &Server{
		uid:           UID(0),
		conn:          conn,
		closed:        make(chan bool),
		requests:      make(chan *Request),
		expiries:      make(chan Expiry),
		notifications: NewCappedPairs(conf.Core.Max),
		display:       NewPopupDisplay(X),
	}
	server.timers = NewUidTimers(server.expiries, conf.Core.Max)
	go server.Loop()
	go server.timers.Loop()
	go xevent.Main(X)

	actionShowAllBind(server).Connect(X, X.RootWin(), conf.Bindings.ShowAll, true)
	actionCloseTopBind(server).Connect(X, X.RootWin(), conf.Bindings.CloseTop, true)
	return server
}

func main() {
	var err error
	conf, err = config.ConfigFromFile("kaze.toml")
	if err != nil {
		panic(err)
	}

	// Parse fonts
	fontBold = mustReadFont(conf.Style.FontBold)
	fontRegular = mustReadFont(conf.Style.FontRegular)
	fontFallback = mustReadFont(conf.Style.FontFallback)

	// DBus handshake
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	reply, err := conn.RequestName("org.freedesktop.Notifications", dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		panic("Name already taken")
	}
	service := NewService(conn, newServer(conn))
	err = conn.Export(service, "/org/freedesktop/Notifications", "org.freedesktop.Notifications")
	if err != nil {
		panic(err)
	}
	select {}
}
