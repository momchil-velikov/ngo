package typ

type (
	S struct{ X, Y float32 }
	T struct{ X, Y float32 }
)

var (
	v0 S = T{1.1, 2.2}
)
