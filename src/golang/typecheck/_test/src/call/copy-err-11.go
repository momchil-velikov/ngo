package call

func f() (int, int, int)

var (
	a = copy(f())
)
