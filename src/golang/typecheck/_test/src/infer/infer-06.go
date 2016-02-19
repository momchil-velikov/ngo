package infer

var s struct {
	X [1]int
	Y [len(s.X)]int
}
