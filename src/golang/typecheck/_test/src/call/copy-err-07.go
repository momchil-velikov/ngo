package call

type (
	Buf   []int
	Buf32 []int32
)

var (
	b0 Buf
	b1 Buf32
	a  = copy(b1, b0)
)
