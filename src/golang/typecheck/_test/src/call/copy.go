package call

type (
	BufT []byte
	BufZ []byte
)

const P = "/tmp"

var (
	src, dst []byte
	a        = copy(dst, src)
	b        = copy(dst, "text")
	c        = copy(dst, P+string("text"))
	buf0     BufT
	d        = copy(buf0, src)
	buf1     BufZ
	e        = copy(buf1, buf0)
)

func f(a []byte) ([]byte, []byte) {
	return a, []byte("abc")
}

func g(a []byte) ([]byte, string) {
	return a, "abc"
}

var (
	h = copy(f(dst))
	i = copy(g(dst))
)
