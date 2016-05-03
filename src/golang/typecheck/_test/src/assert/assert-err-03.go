package assert

type I interface {
	F()
}

type S struct{}

func (*S) F() {}

var (
	A I
	B = A.(S)
)
