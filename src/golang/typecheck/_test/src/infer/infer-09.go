package infer

var a struct {
	X [len(b.U)]int
	Y [1]int
}

var b struct {
	U [1]int
	V [len(a.X)]int
}
