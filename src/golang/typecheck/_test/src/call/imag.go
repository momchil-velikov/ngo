package call

const (
	a = imag(1)
	b = imag(1.0)
	c = imag(1.0 + 0.0i)
	d = imag('b' - 'a')

	c32 = complex(float32(1.0), 1.125)
	c64 = complex(1.0, float64(1.125))

	e = imag(c32)
	f = imag(c64)
)

var (
	x0 complex64  = 1.0 + 1.1i
	x1 complex128 = 1.0 + 1.1i

	x = imag(x0)
	y = imag(x1)
)
