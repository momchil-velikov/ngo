package e

import "a"

func F() int {
	return a.B
}

func G() int {
	return a.A(1)
}

func H() {
	a.A.F
}
