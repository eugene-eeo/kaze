package tctx

import "time"

var ctx *tctx

func init() {
	ctx = &tctx{
		id:       0,
		timer:    nil,
		reqs:     &pairHeap{[]pair{}},
		doneChan: make(chan uint),
		idChan:   make(chan uint),
		reqChan:  make(chan time.Duration),
	}
	go ctx.Loop()
}

func Request(d time.Duration) uint {
	if d < 0 {
		return 0
	}
	return ctx.GetUid(d)
}

func Listen() <-chan uint {
	return ctx.doneChan
}
