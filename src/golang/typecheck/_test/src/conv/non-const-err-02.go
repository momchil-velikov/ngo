package conv

type Int int

var (
	e *int64
	f = (*Int)(e)
)
