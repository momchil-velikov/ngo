package scanner

type lineExtent struct{ off, len int }

// SourceMap maintains a mapping from file offsets to <line, col> pairs.
type SourceMap struct {
	line []lineExtent
	size int
}

func (s *SourceMap) AddLine(sz int) {
	s.line = append(s.line, lineExtent{s.size, sz})
	s.size += sz
}

// Returns the size of the source in bytes.
func (s *SourceMap) Size() int {
	return s.size
}

// Returns the number of source lines.
func (s *SourceMap) LineCount() int {
	return len(s.line)
}

// Returns the start offset and the length of a source line.
func (s *SourceMap) LineExtent(i int) (int, int) {
	return s.line[i].off, s.line[i].len
}

// Returns the 1-based line and column numbers corresponding to the given
// source byte offset.
func (s *SourceMap) Position(off int) (int, int) {
	i := findContainingLine(s.line, off)
	if i < 0 {
		return 0, 0
	} else {
		return i + 1, off - s.line[i].off + 1
	}
}

func findContainingLine(a []lineExtent, off int) int {
	lo, hi := 0, len(a)
	for lo < hi {
		m := lo + (hi-lo)/2
		if a[m].off <= off && off < a[m].off+a[m].len {
			return m
		}
		if a[m].off > off {
			hi = m
		} else {
			lo = m + 1
		}
	}
	return -1
}
