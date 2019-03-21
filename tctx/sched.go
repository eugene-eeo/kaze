package tctx

import "time"
import "container/heap"

type pair struct {
	id uint
	te time.Time
}

type tctx struct {
	id       uint
	timer    *time.Timer
	reqs     *pairHeap
	doneChan chan uint
	idChan   chan uint
	reqChan  chan time.Duration
}

func (tc *tctx) GetUid(d time.Duration) uint {
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
	if tc.timer != nil {
		if !tc.timer.Stop() {
			<-tc.timer.C
		}
	}
	tc.timer = time.NewTimer(0)
	tc.idChan <- tc.id
}

func (tc *tctx) handleTimeout(t time.Time) {
	tc.timer = nil
	for tc.reqs.Len() > 0 {
		p := (*tc.reqs)[0]
		if p.te.Before(t) || p.te.Equal(t) {
			tc.doneChan <- p.id
			heap.Pop(tc.reqs)
		} else {
			m := p.te.Sub(t)
			tc.timer = time.NewTimer(m)
			return
		}
	}
}

func (tc *tctx) Loop() {
	for {
		if tc.timer == nil {
			tc.handleRequest(<-tc.reqChan)
		} else {
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
}
