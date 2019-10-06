package fop


import (
"errors"
"io"
"os"
)

// NewSectionWriter returns a SectionWriter that reads from r
// starting at offset off and stops with EOF after n bytes.
func NewSectionWriter(r io.WriterAt, off int64, n int64) *SectionWriter {
	return &SectionWriter{r, off, off, off + n}
}

// SectionWriter implements Read, Seek, and ReadAt on a section
// of an underlying WriterAt.
type SectionWriter struct {
	r     io.WriterAt
	base  int64
	off   int64
	limit int64
}

func (s *SectionWriter) Write(p []byte) (n int, err error) {
	if s.off >= s.limit {
		return 0, io.EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err = s.r.WriteAt(p, s.off)
	s.off += int64(n)
	return
}

var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

func (s *SectionWriter) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case os.SEEK_SET:
		offset += s.base
	case os.SEEK_CUR:
		offset += s.off
	case os.SEEK_END:
		offset += s.limit
	}
	if offset < s.base {
		return 0, errOffset
	}
	s.off = offset
	return offset - s.base, nil
}

func (s *SectionWriter) WriteAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off >= s.limit - s.base {
		return 0, io.EOF
	}
	off += s.base
	if max := s.limit - off; int64(len(p)) > max {
		p = p[0:max]
		n, err = s.r.WriteAt(p, off)
		if err == nil {
			err = io.EOF
		}
		return n, err
	}
	return s.r.WriteAt(p, off)
}

// Size returns the size of the section in bytes.
func (s *SectionWriter) Size() int64 {
	return s.limit - s.base
}
