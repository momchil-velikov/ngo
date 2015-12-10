package ok

func F(n uint) uint {
	switch n {
	case 0:
	f:
		fallthrough
	case 1:
		return 1
	default:
		return F(n-2) + F(n-1)
	}
}
