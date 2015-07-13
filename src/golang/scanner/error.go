package scanner

import "fmt"

type scanError struct {
	Name      string
	Line, Pos int
	Msg       string
	Sample    string
}

func (e *scanError) Error() string {
	s := fmt.Sprintf("scanner error: %s:%d:%d: ", e.Name, e.Line, e.Pos)
	if len(e.Msg) > 0 {
		s = s + e.Msg
	}
	if len(e.Sample) > 0 {
		s = s + "\n" + e.Sample

	}
	return s
}

func makeError(name string, line, pos int, msg string, sample string) error {
	return &scanError{name, line, pos, msg, sample}
}
