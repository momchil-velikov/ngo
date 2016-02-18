package typ

type I interface{}

type J I

type S struct {
	*J
}
