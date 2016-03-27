package binary

type S struct {
	X [1]int
}

var (
	A, B S
	C    = A < B
)
