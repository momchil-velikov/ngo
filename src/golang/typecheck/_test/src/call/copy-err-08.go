package call

type (
	A  int
	SA []A
	B  int
	SB []B
)

var (
	a SA
	b SB
	c = copy(a, b)
)
