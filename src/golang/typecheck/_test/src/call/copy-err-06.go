package call

type (
	Buf   []int
	Buf32 []int32
)

var (
	b0 Buf
	b1 Buf32
	a  = copy(b0, b1)
)
