package main

import "container/heap"

type pairHeap []*UidPair

func (h pairHeap) Len() int      { return len(h) }
func (h pairHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h pairHeap) Less(i, j int) bool {
	ua := h[i].Notification.Hints.Urgency
	ub := h[j].Notification.Hints.Urgency
	if ua == ub {
		return h[i].Uid < h[j].Uid
	} else {
		return ua < ub
	}
}

func (h *pairHeap) Push(x interface{}) {
	*h = append(*h, x.(*UidPair))
}

func (h *pairHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type CappedPairs struct {
	max   int
	lru   *pairHeap
	pairs map[uint32]*UidPair
}

func NewCappedPairs(max int) *CappedPairs {
	return &CappedPairs{
		max:   max,
		lru:   &pairHeap{},
		pairs: map[uint32]*UidPair{},
	}
}

func (cp *CappedPairs) Insert(x uint32, p *UidPair) (excess *UidPair) {
	heap.Push(cp.lru, p)
	cp.pairs[x] = p
	if cp.max != -1 && cp.lru.Len() > cp.max {
		excess = heap.Pop(cp.lru).(*UidPair)
	}
	return
}

func (cp *CappedPairs) Get(x uint32) *UidPair {
	return cp.pairs[x]
}

func (cp *CappedPairs) Delete(x uint32) {
	pair := cp.pairs[x]
	last := cp.lru.Len() - 1
	for i := last; i >= 0; i-- {
		if (*cp.lru)[i] == pair {
			// swap last and this
			cp.lru.Swap(i, last)
			cp.lru.Pop()
			if i != last {
				heap.Fix(cp.lru, i)
			}
			break
		}
	}
	delete(cp.pairs, x)
}
