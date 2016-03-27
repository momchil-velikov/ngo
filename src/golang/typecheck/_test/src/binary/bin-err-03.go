package binary

type Int int32
type Int32 int32

var (
	A Int
	B Int32
	C = A + B
)
