package main

import "sort"

func lessNotifications(a *Notification, b *Notification) bool {
	ua := a.Hints.Urgency
	ub := a.Hints.Urgency
	if ua == ub {
		return a.Id <= b.Id
	}
	return ua <= ub
}

type pairArray []*UidPair

func (h pairArray) Len() int {
	return len(h)
}

func (h *pairArray) Find(p *UidPair) int {
	hh := *h
	n := p.Notification
	return sort.Search(h.Len(), func(i int) bool {
		if hh[i] == p {
			return true
		}
		a := hh[i].Notification
		ua := a.Hints.Urgency
		ub := n.Hints.Urgency
		if ua == ub {
			return a.Id <= n.Id
		}
		return ua <= ub
	})
}

func (h *pairArray) Delete(p *UidPair) {
	hh := *h
	i := h.Find(p)
	n := len(hh)
	copy(hh[i:], hh[i+1:])
	hh[n-1] = nil // or the zero value of T
	hh = hh[:n-1]
	*h = hh
}

func (h *pairArray) Insert(p *UidPair) {
	hh := *h
	i := h.Find(p)
	hh = append(hh, nil)
	copy(hh[i+1:], hh[i:])
	hh[i] = p
	*h = hh
}

type CappedPairs struct {
	max   int
	lru   *pairArray
	pairs map[uint32]*UidPair
}

func NewCappedPairs(max int) *CappedPairs {
	lru := make(pairArray, 0, max+1)
	return &CappedPairs{
		max:   max,
		lru:   &lru,
		pairs: make(map[uint32]*UidPair, max+1),
	}
}

func (cp *CappedPairs) Insert(x uint32, p *UidPair) (excess *UidPair) {
	cp.pairs[x] = p
	cp.lru.Insert(p)
	if cp.max > 0 && cp.lru.Len() > cp.max {
		excess = (*cp.lru)[cp.lru.Len()-1]
	}
	return
}

func (cp *CappedPairs) Get(x uint32) *UidPair {
	return cp.pairs[x]
}

func (cp *CappedPairs) Delete(x uint32) {
	if p, ok := cp.pairs[x]; ok {
		delete(cp.pairs, x)
		cp.lru.Delete(p)
	}
}
