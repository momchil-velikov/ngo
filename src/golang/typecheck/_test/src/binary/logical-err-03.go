package binary

type Bool bool

const (
	A bool = true
	B Bool = false
	C      = A && B
)
