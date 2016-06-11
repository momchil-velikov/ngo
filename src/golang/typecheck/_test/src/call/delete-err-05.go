package call

var (
	m map[int8]string
	s = uint(1)
	a = delete(m, 0x80>>s)
)
