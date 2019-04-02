package main

import "time"
import "github.com/godbus/dbus"

type UID uint

type UidPair struct {
	Uid          UID
	Notification *Notification
}

type Server struct {
	uid           UID
	conn          *dbus.Conn
	closed        chan bool // used for CloseNotification
	requests      chan *Request
	expiries      chan Expiry
	notifications *CappedPairs
	timers        *UidTimers
	display       *PopupDisplay
}

func (s *Server) calculateTimeouts(u Urgency, expireTimeout time.Duration) (expiry time.Duration, popup time.Duration) {
	popupAge := conf.Core.MaxPopupAge.Duration
	timeout := expireTimeout
	if u == UrgencyCritical {
		popupAge = -1
	}
	return popupAge, timeout
}

func (s *Server) close(uid UID, id uint32, reason Reason) {
	s.notifications.Delete(uid)
	s.display.Destroy(uid)
	s.timers.Delete(uid)
	s.emitClosed(id, reason)
}

func (s *Server) redraw() {
	s.display.Draw(s.notifications.Order())
}

func (s *Server) advanceUid() {
	s.uid++
	if s.uid == 0 {
		s.uid++
	}
}

func (s *Server) handleNewNotification(n *Notification) {
	s.advanceUid()
	old := s.uid
	new := s.uid
	// if this is a replacesId call
	if old_id := s.notifications.UidOf(n.Id); old_id > 0 {
		old = old_id
		s.timers.Delete(old_id)
	}
	excess := s.notifications.Insert(new, n)
	if excess != nil {
		s.close(excess.Uid, excess.Notification.Id, ReasonUndefined)
	}
	// If we have no excess, or we are NOT the excess, then display it
	if excess == nil || excess.Uid != new {
		s.display.Show(old, new, n, actionContextMenuCb(s), actionCloseOneCb(s))
		// calculate and add timeouts
		popupAge, timeout := s.calculateTimeouts(n.Hints.Urgency, n.ExpireTimeout)
		s.timers.Add(popupAge, Expiry{ExpiryPopupClose, new})
		s.timers.Add(timeout, Expiry{ExpiryTimeout, new})
		s.redraw()
	}
}

func (s *Server) emitAction(id uint32, action_key string) {
	s.conn.Emit("/org/freedesktop/Notifications", "org.freedesktop.Notifications.ActionInvoked", id, action_key)
}

func (s *Server) emitClosed(id uint32, reason Reason) {
	s.conn.Emit("/org/freedesktop/Notifications", "org.freedesktop.Notifications.NotificationClosed", id, reason)
}

func (s *Server) handleExpiry(exp Expiry) {
	noti := s.notifications.Get(exp.Uid)
	if noti != nil {
		if exp.Type == ExpiryTimeout {
			s.close(exp.Uid, noti.Id, ReasonExpired)
		} else {
			s.display.Close(exp.Uid)
		}
		s.redraw()
	}
}

func (s *Server) handleCloseNotification(id uint32) bool {
	uid := s.notifications.UidOf(id)
	if uid == 0 {
		return false
	}
	s.close(uid, id, ReasonCloseNotification)
	return true
}

func (s *Server) handleAction(a ActionRequest) {
	switch a.Type {
	case ActionCloseOne:
		// why a.Uid is not taken into account:
		// the user should never be allowed to close the notification before an
		// update (replacesId) completes
		nid := a.Nid
		uid := s.notifications.UidOf(nid)
		if uid != 0 {
			s.close(uid, nid, ReasonUserDismissed)
			s.redraw()
		}
	case ActionCloseTop:
		pair := s.display.FirstVisible(s.notifications.Order())
		if pair != nil {
			s.close(pair.Uid, pair.Notification.Id, ReasonUserDismissed)
		}
	case ActionContextMenu:
		nid := a.Nid
		uid := s.notifications.UidOf(nid)
		noti := s.notifications.Get(uid)
		if noti == nil {
			return
		}
		go execMixedSelector(noti, func(action string) {
			// If there are no actions/links we will get action == ""
			if action != "" {
				s.emitAction(nid, action)
			}
			if !noti.Hints.Resident {
				s.requests <- &Request{
					Type: RequestAction,
					Body: ActionRequest{ActionContextMenuDone, nid, uid},
				}
			}
		})
	case ActionContextMenuDone:
		if s.notifications.UidOf(a.Nid) == a.Uid {
			s.close(a.Uid, a.Nid, ReasonUndefined)
			s.redraw()
		}
	case ActionShowAll:
		for _, u := range s.notifications.Order() {
			s.display.Show(u.Uid, u.Uid, u.Notification, actionContextMenuCb(s), actionCloseOneCb(s))
			if u.Notification.Hints.Urgency != UrgencyCritical {
				s.timers.Add(conf.Core.MaxPopupAge.Duration, Expiry{
					ExpiryPopupClose,
					u.Uid,
				})
			}
		}
		s.redraw()
	}
}

func (s *Server) Loop() {
	for {
		select {
		case req := <-s.requests:
			switch req.Type {
			case RequestNewNotification:
				s.handleNewNotification(req.Body.(*Notification))
			case RequestCloseNotification:
				s.closed <- s.handleCloseNotification(req.Body.(uint32))
			case RequestAction:
				s.handleAction(req.Body.(ActionRequest))
			}
		case exp := <-s.expiries:
			s.handleExpiry(exp)
		}
	}
}

func (s *Server) HandleNotification(n *Notification) {
	s.requests <- &Request{RequestNewNotification, n}
}

func (s *Server) HandleClose(id uint32) *dbus.Error {
	s.requests <- &Request{RequestCloseNotification, id}
	if !<-s.closed {
		return dbus.NewError("CloseNotification", nil)
	}
	return nil
}
