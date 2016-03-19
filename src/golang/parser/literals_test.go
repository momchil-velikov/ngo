package parser

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
		if string(s) != out[i] {
			t.Errorf("error converting input %v: %v", in[i], []byte(s))
		}
	}
}

func TestRune(t *testing.T) {
	r := Rune([]byte("め"))
	if r.Int64() != 'め' {
		t.Error("rune decode error")
	}
}

func TestInt(t *testing.T) {
	for _, e := range []struct {
		in  string
		out string
	}{
		{"12", "12"},
		{"012", "10"},
		{"0x12", "18"},
		{"0x1234567890abcdef12", "335812727627494321938"},
		{"4865125723542547825342548254127854", "4865125723542547825342548254127854"},
	} {
		x := Int([]byte(e.in))
		if x.String() != e.out {
			t.Errorf("unexpected value `%s`, should be `%s`", x.String(), e.out)
		}
	}
}

func TestFloat(t *testing.T) {
	for _, e := range []struct {
		in  string
		out string
	}{
		{"1.125", "1.125"},
		{"012", "12"},
		{"0x12", "18"},
		{
			"0x123456789abcdef0123456789abcdef0123456789abcdef0123456789abc.def0p+16",
			"8234104123542484900769178205574010627627573691361805720124810878238590820080",
		},
		{"0x1234567890abcdef12", "335812727627494321938"},
		{"4865125723542547825342548254127854", "4865125723542547825342548254127854"},

		{
			"3.1415926535897932384626433832795028841971693993751058209749445923078164062862",

			"3.1415926535897932384626433832795028841971693993751058209749445923078164062862"},
	} {
		x, err := Float([]byte(e.in))
		if err != nil {
			t.Fatal(err)
		}
		out := x.Text('g', 78)
		if out != e.out {
			t.Errorf("unexpected value `%s`, should be `%s`", out, e.out)
		}
	}
}
