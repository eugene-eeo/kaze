package tctx

import "time"

var ctx *tctx

func init() {
	ctx = &tctx{
		id:       0,
		timer:    time.NewTimer(0),
		reqs:     &pairHeap{},
		doneChan: make(chan TimerId),
		idChan:   make(chan TimerId),
		reqChan:  make(chan time.Duration),
	}
	go ctx.Loop()
}

func Request(d time.Duration) TimerId {
	return ctx.GetUid(d)
}

func Listen() <-chan TimerId {
	return ctx.doneChan
}
