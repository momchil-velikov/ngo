package constexpr

import "testing"

func isSame(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestString(t *testing.T) {
	in := [][]byte{
		[]byte{'"', 'a', '"'},
		[]byte{'"', 'a', '\\', '0', '0', '1', '"'},
		[]byte{'"', 'a', '\\', 'x', '0', '1', '"'},
		[]byte{'"', 'a', '\\', 'u', '0', 'a', '2', '3', '4', '"'},
		[]byte{'"', '\\', 'U', '0', '0', '0', '2', '1', '6', '8', 'B', 'e', '"'},
		[]byte{'"', '\\', 'a', '0', '\\', 'b', '\\', 'f', '6', '\\', 'n',
			'\\', 'r', '\\', 't', '\\', 'v', '\\', '\\', '\'', '\\', '"', '"'},
		[]byte{'\'', '\'', 'a', '"', '\n', '\\', 'x', 'a', 'D',
			'\''},
	}
	out := []string{"a", "a\001", "a\x01", "a\u0a234", "\U0002168Be",
		"\a0\b\f6\n\r\t\v\\'\"", "'a\"\\xaD",
	}

	for i := range in {
		s := String(in[i])
		if s != out[i] {
			t.Errorf("error converting input %v: %v", in[i], []byte(s))
		}
	}
}
