package infer

var u = s.X
var v = s.Y

type S struct {
	X [1]int
	Y [len(u)]int
}

var s S
