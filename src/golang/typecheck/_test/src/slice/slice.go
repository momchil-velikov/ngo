package slice

type Int int
type T [5]Int
type S []Int

var (
	a string
	b = a[:]
	c = a[:2]
	d = a[1:]
	e = "hello"[5:5]

	p *T
	q T
	f = p[:]
	g = q[:]
	h = p[:2]
	i = q[1:]
	j = p[5:5]
	k = q[5:5:5]
	l = p[:5:5]

	qq S
	ff = qq[:]
	gg = qq[:]
	hh = qq[:2]
	ii = qq[1:]
	jj = qq[5:5]
	kk = qq[5:5:5]
	ll = qq[:5:5]
)
