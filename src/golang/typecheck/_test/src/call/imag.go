package call

const (
	a = imag(1)
	b = imag(1.0)
	c = imag(1.0 + 0.0i)
	d = imag('b' - 'a')

	// FIXME: add more tests after implementing `complex``
)

var (
	x0 complex64  = 1.0 + 1.1i
	x1 complex128 = 1.0 + 1.1i

	x = imag(x0)
	y = imag(x1)
)
