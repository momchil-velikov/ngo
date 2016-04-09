package comp

type S struct {
	X, Y int
}

type T struct {
	S
	Z int
}

var (
	a = S{}
	b = S{a.Y, a.X}
	c = T{S{1, 2}, 3}
	d = T{S: S{1, 2}, Z: c.Z}
	e = S{X: 1}
)
