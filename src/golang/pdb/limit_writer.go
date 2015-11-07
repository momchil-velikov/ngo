package pdb

import (
	"errors"
	"io"
)

var ErrorNoSpace = errors.New("no space left on device")

type LimitedWriter struct {
	N int
	W io.Writer
}

func (w *LimitedWriter) Write(b []byte) (int, error) {
	if len(b) > w.N {
		b = b[:w.N]
		n, err := w.W.Write(b)
		if err == nil {
			w.N -= n
			if w.N == 0 {
				err = ErrorNoSpace
			}
		}
		return n, err
	}
	n, err := w.W.Write(b)
	if err == nil {
		w.N -= n
	}
	return n, err
}

var _ io.Writer = &LimitedWriter{}

func LimitWriter(w io.Writer, n int) *LimitedWriter {
	return &LimitedWriter{N: n, W: w}
}
