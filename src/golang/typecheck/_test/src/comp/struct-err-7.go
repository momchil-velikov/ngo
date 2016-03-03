package comp

type S struct {
	X, Y int
}

type T struct {
	S
	Z int
}

var a = T{S: {1, 2}, Z: 3}
