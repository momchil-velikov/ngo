package call

type (
	Int8   int8
	String string
)

func f() (Int8, Int8)

var (
	a = delete(f())
)
