package sel

import "sel/b"

var (
	z b.C

	w = z.y // error
)
