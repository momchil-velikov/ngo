package ok

const (
	A = 1
	B
	C, D int = A, B
	E, F
)

func Fn() {
	const (
		A = 1
		B
		C, D int = A, B
		E, F
	)
}
