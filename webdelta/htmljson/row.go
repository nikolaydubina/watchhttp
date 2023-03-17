package htmljson

import (
	"io"
	"strings"
)

// RowWriter accumulates items written to row and flushes it as a wrapped row on flush calls
// flush has to be called eventually.
type RowWriter struct {
	b   strings.Builder
	w   io.Writer
	Row func(s string, padding int) string
	err []error
}

func (s *RowWriter) write(v string) {
	_, err := s.b.WriteString(v)
	s.err = append(s.err, err)
}

func (s *RowWriter) flush(depth int) {
	v := s.Row(s.b.String()+"\n", depth)
	_, err := io.WriteString(s.w, v)
	s.err = append(s.err, err)
	s.b.Reset()
}
