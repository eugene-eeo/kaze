package tctx

type pairHeap []pair

func (p pairHeap) Len() int           { return len(p) }
func (p pairHeap) Less(i, j int) bool { return p[i].te.Before(p[j].te) }
func (p pairHeap) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p *pairHeap) Push(x interface{}) {
	*p = append(*p, x.(pair))
}

func (p *pairHeap) Pop() interface{} {
	old := *p
	n := len(old)
	x := old[n-1]
	*p = old[0 : n-1]
	return x
}
