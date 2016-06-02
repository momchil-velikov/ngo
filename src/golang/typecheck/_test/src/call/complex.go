package call

const (
	a = complex(1, 2)
	b = complex(1.0, 2.0)
	c = complex('b'-'a', 'c'-'a')

	d = complex(float32(1), 2)
	e = complex(1, float64(2))

	ff = complex(complex(1.0, 0.0), complex(2.0, 0.0))
)

func g32() (float32, float32)
func g64() (float64, float64)

var (
	s, t float32
	u, v float64

	f = complex(s, 2.0)
	g = complex(1.0, t)
	h = complex(s, t)
	x = complex(u, 2.0)
	y = complex(1.0, v)
	z = complex(u, v)

	p = complex(g32())
	q = complex(g64())
)
