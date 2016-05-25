package call

func f() int

var (
	ss []int
	a  = f(ss...)
)
