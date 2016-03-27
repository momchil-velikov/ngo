package binary

type I interface {
	F()
}

type S struct {
	X []int
}

func (S) F()

var (
	A I
	B S
	C = B == A
)
