package call

type (
	Int8   int8
	String string
)

func f() (map[Int8]String, Int8, Int8)

var (
	a = delete(f())
)
