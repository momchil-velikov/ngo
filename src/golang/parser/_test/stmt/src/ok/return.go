package ok

func F() {
	return
}

var A, B int

func G() (int, int) {
	return A, B
}

func H() {
	break
	continue
	goto L
	fallthrough
}
