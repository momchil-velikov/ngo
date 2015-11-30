package ok

var A, B int

func F() {
	A = 1
	A, B = B, A
}
