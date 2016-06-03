package call

const (
	a = real(1)
	b = real(1.0)
	c = real(1.0 + 1.1i)
	d = real('b' - 'a')

	c32 = complex(float32(1.0), 1.1)
	c64 = complex(1.0, float64(1.1))

	e = real(c32)
	f = real(c64)
)

var (
	x0 complex64  = 1.0 + 1.1i
	x1 complex128 = 1.0 + 1.1i

	x = real(x0)
	y = real(x1)
)
