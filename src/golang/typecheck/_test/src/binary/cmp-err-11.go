package binary

type I interface {
	F()
}

type S struct{}

var (
	A I
	B S
	C = B == A
)
