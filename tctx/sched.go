package tctx

import "time"
import "container/heap"

type TimerId uint

type pair struct {
	id TimerId
	te time.Time
}

type tctx struct {
	id       TimerId
	timer    *time.Timer
	reqs     *pairHeap
	doneChan chan TimerId
	idChan   chan TimerId
	reqChan  chan time.Duration
}

func (tc *tctx) GetUid(d time.Duration) TimerId {
	// Sentinel value for timers that never fire
	if d < 0 {
		return 0
	}
	tc.reqChan <- d
	return <-tc.idChan
}

func (tc *tctx) handleRequest(d time.Duration) {
	now := time.Now()
	tc.id++
	if tc.id == 0 {
		tc.id++
	}
	heap.Push(tc.reqs, pair{tc.id, now.Add(d)})
	tc.timer.Reset(0)
	tc.idChan <- tc.id
}

func (tc *tctx) handleTimeout(t time.Time) {
	for tc.reqs.Len() > 0 {
		p := (*tc.reqs)[0]
		m := p.te.Sub(t)
		if m <= 0 {
			tc.doneChan <- p.id
			heap.Pop(tc.reqs)
		} else {
			tc.timer.Reset(m)
			return
		}
	}
}

func (tc *tctx) Loop() {
	for {
		select {
		// new requests
		case d := <-tc.reqChan:
			tc.handleRequest(d)
		// for timeouts
		case t := <-tc.timer.C:
			tc.handleTimeout(t)
		}
	}
}
