package call

const (
	a = real(1)
	b = real(1.0)
	c = real(1.0 + 1.1i)
	d = real('b' - 'a')

	// FIXME: add more tests after implementing `complex``
)

var (
	x0 complex64  = 1.0 + 1.1i
	x1 complex128 = 1.0 + 1.1i

	x = real(x0)
	y = real(x1)
)
