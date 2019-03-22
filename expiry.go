package main

import "time"
import "github.com/eugene-eeo/kaze/tctx"

type TimerInfo struct {
	TimeoutId    tctx.TimerId
	PopupCloseId tctx.TimerId
}

type UidTimers struct {
	e map[tctx.TimerId]Expiry
	m map[UID]*TimerInfo
	c chan Expiry
}

func NewUidTimers(c chan Expiry) *UidTimers {
	return &UidTimers{
		e: map[tctx.TimerId]Expiry{},
		m: map[UID]*TimerInfo{},
		c: c,
	}
}

func (ut *UidTimers) Delete(uid UID) {
	delete(ut.m, uid)
}

func (ut *UidTimers) Add(d time.Duration, e Expiry) {
	timerId := tctx.Request(d)
	ut.e[timerId] = e
	if ut.m[e.Uid] == nil {
		ut.m[e.Uid] = &TimerInfo{}
	}
	switch e.Type {
	case ExpiryTimeout:
		ut.m[e.Uid].TimeoutId = timerId
	case ExpiryPopupClose:
		ut.m[e.Uid].PopupCloseId = timerId
	}
}

func (ut *UidTimers) Loop() {
	for {
		timerId := <-tctx.Listen()
		expiry, ok := ut.e[timerId]
		if !ok {
			continue
		}
		info := ut.m[expiry.Uid]
		if info != nil && (info.TimeoutId == timerId || info.PopupCloseId == timerId) {
			delete(ut.m, expiry.Uid)
			ut.c <- expiry
		}
	}
}
