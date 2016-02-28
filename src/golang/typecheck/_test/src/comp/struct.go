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
	b = S{1, 2}
	c = T{S{1, 2}, 3}
	d = T{S: S{1, 2}, Z: 3}
	e = S{X: 1}
)
