package conv

type Int int

var (
	e *Int
	f = (*int64)(e)
)
