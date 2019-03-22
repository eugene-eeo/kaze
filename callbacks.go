package main

import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xevent"
import "github.com/BurntSushi/xgbutil/keybind"

func actionCloseOneCb(s *Server) func(n *Notification) {
	return func(n *Notification) {
		s.requests <- &Request{
			Type: RequestAction,
			Body: ActionRequest{
				Type: ActionCloseOne,
				Nid:  n.Id,
			},
		}
	}
}

func actionContextMenuCb(s *Server) func(n *Notification) {
	return func(n *Notification) {
		s.requests <- &Request{
			Type: RequestAction,
			Body: ActionRequest{
				Type: ActionContextMenu,
				Nid:  n.Id,
			},
		}
	}
}

func actionShowAllBind(s *Server) keybind.KeyPressFun {
	return keybind.KeyPressFun(func(*xgbutil.XUtil, xevent.KeyPressEvent) {
		s.requests <- &Request{
			Type: RequestAction,
			Body: ActionRequest{Type: ActionShowAll},
		}
	})
}

func actionCloseTopBind(s *Server) keybind.KeyPressFun {
	return keybind.KeyPressFun(func(*xgbutil.XUtil, xevent.KeyPressEvent) {
		s.requests <- &Request{
			Type: RequestAction,
			Body: ActionRequest{Type: ActionCloseTop},
		}
	})
}
