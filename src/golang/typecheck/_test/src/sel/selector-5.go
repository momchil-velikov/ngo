package sel

import "sel/b"

var (
	y b.B

	v = y.y // error
)
