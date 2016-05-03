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

const (
	X = -1
	Y = iota
	Z
)

const (
	AA = iota
	BB = iota + DD
	CC
)
const (
	DD = iota + 1
)
