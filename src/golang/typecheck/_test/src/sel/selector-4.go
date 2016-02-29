package sel

import "sel/a"

var (
	x a.A

	v = x.y // error
)
