package conv

type Int int
type Byte byte
type Rune rune
type String string

type I interface{}

type J interface {
	I
}
type K interface {
	I
}
type EmptyIf interface {
	J
	K
}

type NonEmptyIf interface {
	EmptyIf
	F()
}

const C = 1

var (
	// "assignable"
	a int
	b = int(a)

	// identical underlying types
	c Int = Int(a)
	d     = int(c)

	// unnamed pointers
	e *Int
	f = (*int)(e)
	g = (*Int)(f)

	// integer -> string
	i int
	h = String(i)

	// []byte -> string
	j []Byte
	k = String(j)

	// []rune -> string
	l []Rune
	m = String(j)

	// string -> []byte
	n = []Byte(m)

	// string -> []rune
	o = []Rune(m)

	// between integers
	p int16
	q = uint32(p)
	r = int8(q)

	// betweeen floats
	s float32
	t = float64(s)
	u = float32(t)

	// between complex
	v complex64
	w = complex128(v)
	x = complex64(w)

	// constant value representable by `interface{}`
	if0, if1         = EmptyIf(1), EmptyIf(C)
	if2, if3 EmptyIf = 1.1, C + 1.1
)
