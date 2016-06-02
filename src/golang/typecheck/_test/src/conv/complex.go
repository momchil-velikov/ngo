package conv

const (
	a = int(complex(1.0, 0.0))
	b = float32(complex(1.0, 0.0))
	u = float32(1.0)
	c = float64(complex(u, 0.0))

	cc = complex(0, 3.4028235e+38)
	d  = complex64(cc)

	ccc = complex(0, 3.4028236e+38)
	dd  = complex128(cc)
)
