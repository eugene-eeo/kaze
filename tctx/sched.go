package tctx

import "time"

type pair struct {
	id uint
	rt time.Duration
}

type tctx struct {
	id       uint
	timer    *time.Timer
	reqs     []*pair
	doneChan chan uint
	idChan   chan uint
	reqChan  chan time.Duration
	tprev    time.Time
}

func (tc *tctx) GetUid(d time.Duration) uint {
	tc.reqChan <- d
	return <-tc.idChan
}

func (tc *tctx) handleRequest(d time.Duration) {
	tc.id++
	if tc.id == 0 {
		tc.id++
	}
	tc.reqs = append(tc.reqs, &pair{tc.id, d})
	if tc.timer == nil {
		tc.timer = time.NewTimer(d)
	}
	tc.idChan <- tc.id
}

func (tc *tctx) handleTimeout(t time.Time) {
	n := 0 // number of expired requests
	m := time.Duration(0)
	dt := t.Sub(tc.tprev)
	for i := len(tc.reqs) - 1; i >= 0; i-- {
		r := tc.reqs[i]
		r.rt -= dt
		if r.rt <= 0 {
			tc.doneChan <- r.id
			// delete this element
			copy(tc.reqs[i:], tc.reqs[i+1:])
			tc.reqs[len(tc.reqs)-1] = nil
			tc.reqs = tc.reqs[:len(tc.reqs)-1]
		} else {
			if n == 0 || r.rt < m {
				m = r.rt
			}
			n++
		}
	}
	tc.tprev = t
	tc.timer.Stop()
	tc.timer = nil
	if n > 0 {
		tc.timer = time.NewTimer(m)
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
