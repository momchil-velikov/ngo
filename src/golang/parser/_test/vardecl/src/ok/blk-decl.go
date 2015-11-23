package ok

func Fn(int) (int, int)

func G() {
	var A, B int
	var (
		C, D int = A, B
	)
	var E, F = Fn(A)
	var _, G = Fn(B)
}
