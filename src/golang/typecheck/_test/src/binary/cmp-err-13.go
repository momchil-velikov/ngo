package binary

type S struct {
	X []int
}

var (
	A, B S
	C    = A != B
)
