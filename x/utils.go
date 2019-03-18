package x

func maxWidth(text string, max int, oracle func(string) int) string {
	n := len(text)
	l := oracle(text)
	i := n
	// minimisation stage
	for l > max && i > 0 {
		i /= 2
		l = oracle(text[:i])
	}
	// maximisation stage
	for i < n {
		i++
		l = oracle(text[:i])
		if l > max {
			i--
			break
		}
	}
	return text[:i]
}
