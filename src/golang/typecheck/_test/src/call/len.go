package call

type String string

const P = "/tmp/"

var (
	a   [8]int
	p   *[8]int
	s   []float32
	ch  chan int
	mm  map[string]int
	str String

	c0  = len(a)
	c1  = len(p)
	c2  = len(s)
	c3  = len(ch)
	c4  = len([...]int{1, 2, 7: 4})
	c5  = len(&[...]int{1, 6: 2, 4})
	c6  = len([...]int{1, <-ch, 2})
	c7  = len([...]int{1, len(s), 2})
	c8  = len(mm)
	c9  = len(str)
	c10 = len("01234567")
	c11 = len(P + "xyz")
	pfx string
	c12 = len(pfx + "xyz")
)
