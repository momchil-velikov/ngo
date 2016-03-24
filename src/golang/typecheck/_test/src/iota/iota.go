package iota

const (
	A = iota
	B
	C = F
	U = ^uint16(1 << iota)
	V
)

const (
	D = int(iota)
	E
	F
	G = V
)

const N = iota
