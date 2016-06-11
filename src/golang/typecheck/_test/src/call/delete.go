package call

type (
	Int8   int8
	String string
)

func f() (map[Int8]String, Int8)

var (
	m map[Int8]String
	a = delete(m, 1)
	b = delete(f())
)
