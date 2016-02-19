package infer

var a, b = s.X, s.Y

var s struct {
	X [len(b)]int
	Y [len(a)]int
}
