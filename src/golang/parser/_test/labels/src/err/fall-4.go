package err

func F(n uint) {
	switch n {
	case 0:
		n = 1
	default:
		fallthrough
	}
}
