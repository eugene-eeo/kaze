package tctx

type pairHeap struct {
	s []pair
}

func (p *pairHeap) Len() int {
	return len(p.s)
}

func (p *pairHeap) Less(i, j int) bool {
	return p.s[i].te.Before(p.s[j].te)
}

func (p *pairHeap) Swap(i, j int) {
	p.s[i], p.s[j] = p.s[j], p.s[i]
}

func (p *pairHeap) Push(x interface{}) {
	p.s = append(p.s, x.(pair))
}

func (p *pairHeap) Pop() interface{} {
	var x interface{}
	x, p.s = p.s[len(p.s)-1], p.s[:len(p.s)-1]
	return x
}
