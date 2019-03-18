package x

func maxWidth(text string, max int, oracle func(string) int) string {
	l := oracle(text)
	i := len(text)
	// minimisation stage
	for l > max {
		i /= 2
		l = oracle(text[:i])
	}
	// maximisation stage
	for {
		i++
		l = oracle(text[:i])
		if l > max {
			i--
			break
		}
	}
	return text[:i]
}
