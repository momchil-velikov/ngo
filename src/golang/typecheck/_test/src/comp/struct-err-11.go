package comp

type S struct {
	X, Y int8
}

var (
	v int
	a = S{X: v, Y: v}
)
