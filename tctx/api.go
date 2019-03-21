package tctx

import "time"

var ctx *tctx

func init() {
	ctx = &tctx{
		id:       0,
		timer:    time.NewTimer(0),
		reqs:     &pairHeap{},
		doneChan: make(chan uint),
		idChan:   make(chan uint),
		reqChan:  make(chan time.Duration),
	}
	go ctx.Loop()
}

func Request(d time.Duration) uint {
	return ctx.GetUid(d)
}

func Listen() <-chan uint {
	return ctx.doneChan
}
