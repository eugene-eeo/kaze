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

func NewUidTimers(c chan Expiry, size int) *UidTimers {
	if size < 0 {
		size = 10
	}
	return &UidTimers{
		e: make(map[tctx.TimerId]Expiry, size*2),
		m: make(map[UID]*TimerInfo, size),
		c: c,
	}
}

func (ut *UidTimers) Delete(uid UID) {
	if timerInfo := ut.m[uid]; timerInfo != nil {
		delete(ut.e, timerInfo.PopupCloseId)
		delete(ut.e, timerInfo.TimeoutId)
	}
	delete(ut.m, uid)
}

func (ut *UidTimers) Add(d time.Duration, e Expiry) {
	timerId := tctx.Request(d)
	ut.e[timerId] = e
	timerInfo := ut.m[e.Uid]
	if timerInfo == nil {
		timerInfo = &TimerInfo{}
		ut.m[e.Uid] = timerInfo
	}
	switch e.Type {
	case ExpiryTimeout:
		delete(ut.e, timerInfo.TimeoutId)
		timerInfo.TimeoutId = timerId
	case ExpiryPopupClose:
		delete(ut.e, timerInfo.PopupCloseId)
		timerInfo.PopupCloseId = timerId
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
