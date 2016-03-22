package b

import "identity/a"

type G3 struct {
	X a.A0 `a:"b"`
	Y a.A1
}

type G5 struct {
	X a.A0 `a:"b"`
	y a.A1
}

type I5 interface {
	H() a.A1
	a.I3
}

type I7 interface {
	H() a.A1
	a.I3
	gG()
}
