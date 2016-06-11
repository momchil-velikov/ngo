package call

type (
	Int8   int8
	String string
)

func f() (map[Int8]String, int)

var (
	a = delete(f())
)
