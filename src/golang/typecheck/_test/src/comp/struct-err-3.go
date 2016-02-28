package comp

type S struct {
	X, Y int
}

type T struct {
	S
	Z int
}

var a = T{X: 1, Y: 2, Z: 3}
