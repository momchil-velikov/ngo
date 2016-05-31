package call

type S struct{ X, Y float64 }

var (
	s S
	a = imag(s)
)
