package correct

type a bool
type b byte
type c uint8
type d uint16
type e uint32
type f uint64
type g int8
type h int16
type (
	i rune
	j int32
	k int64

	aa  a
	aaa aa
)

type rr r

type kkk kk

func F() {
	type A int
	type AA A
	{
		type (
			B  int
			BB AA
		)
		type BBB BB
	}
}

type u [3]a
type v []u
type w *v
type x map[a]w
type y chan x

type z struct {
	X u
	Y v
}

type Fn func(u, v) (w, x)

type IfA interface {
	F(u) v
}

type IfB interface {
	IfA
	G(x) y
}

type _ int
