package call

type SInt []int

const P = "T_"

var (
	s  []int
	ss SInt
	a  = append(s, 1)
	b  = append(a, 1, 2, 3)
	c  = append(ss, s...)

	buf []byte
	d   = append(buf, "abc"...)
	e   = append(buf, d...)
	f   = append(buf, (P + "abc")...)
	g   = append(buf)
)

func fn() []byte
func gn() ([]byte, byte, byte)

var (
	h = append(fn())
	i = append(gn())
)
