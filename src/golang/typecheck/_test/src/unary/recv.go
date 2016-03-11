package unary

type T int

var (
	in    <-chan T
	inout chan T

	A     = <-in
	B, ok = <-inout
)
