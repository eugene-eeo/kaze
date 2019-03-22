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
	max     int
	lru     *pairArray
	lookup  map[UID]*Notification
	idToUid map[uint32]UID // maps from notification-id to uid
}

func NewCappedPairs(max int) *CappedPairs {
	lru := make(pairArray, 0, max+1)
	return &CappedPairs{
		max:     max,
		lru:     &lru,
		lookup:  map[UID]*Notification{},
		idToUid: map[uint32]UID{}, // maps from notification-id to uid
	}
}

func (cp *CappedPairs) Insert(uid UID, p *Notification) (excess *UidPair) {
	// If this is a replacesId operation then we need to delete the previous
	// notification
	x := cp.idToUid[p.Id]
	if old_noti := cp.lookup[x]; old_noti != nil {
		delete(cp.lookup, x)
		cp.lru.Delete(&UidPair{x, old_noti})
	}
	cp.idToUid[p.Id] = uid
	cp.lookup[uid] = p
	cp.lru.Insert(&UidPair{uid, p})
	if cp.max > 0 && cp.lru.Len() > cp.max {
		excess = (*cp.lru)[cp.lru.Len()-1]
	}
	return
}

func (cp *CappedPairs) UidOf(x uint32) UID {
	return cp.idToUid[x]
}

func (cp *CappedPairs) Get(uid UID) *Notification {
	return cp.lookup[uid]
}

func (cp *CappedPairs) GetByNotificationId(x uint32) *Notification {
	return cp.lookup[cp.idToUid[x]]
}

func (cp *CappedPairs) DeleteByNotificationId(x uint32) {
	cp.Delete(cp.idToUid[x])
}

func (cp *CappedPairs) Delete(uid UID) {
	noti := cp.lookup[uid]
	if noti != nil {
		cp.lru.Delete(&UidPair{uid, noti})
		delete(cp.lookup, uid)
		if cp.idToUid[noti.Id] == uid {
			delete(cp.idToUid, noti.Id)
		}
	}
}

func (cp *CappedPairs) Order() []*UidPair {
	return *cp.lru
}

func (cp *CappedPairs) Top() (uint32, UID) {
	order := cp.Order()
	if len(order) == 0 {
		return 0, 0
	}
	return order[0].Notification.Id, order[0].Uid
}
