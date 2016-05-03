package assert

type I interface {
	F()
}

var (
	A I
	B = A.([]int)
)
