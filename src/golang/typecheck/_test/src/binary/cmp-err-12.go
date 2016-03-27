package binary

type I interface {
	F()
}

type J interface {
	G()
}

var (
	A I
	B J
	C = A != B
)
