package ok

var A, B int

func F() {
	A := A
	A, B := B, A
	C, _ := 1, 2
}
