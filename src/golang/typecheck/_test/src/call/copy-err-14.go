package call

type (
	A int
	B int
)

func f() ([]A, []B)

var (
	a = copy(f())
)
