package binary

var (
	A, B interface {
		F()
	}
	C = A <= B
)
