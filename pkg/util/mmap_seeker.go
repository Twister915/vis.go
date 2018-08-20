package util

import (
	"errors"
	"io"

	"golang.org/x/exp/mmap"
)

type MMapSeeker struct {
	M   *mmap.ReaderAt
	off int64
}

func (m *MMapSeeker) Read(to []byte) (n int, err error) {
	stop := m.off + int64(len(to))
	l := int64(m.M.Len())
	if stop > l {
		stop = l
	}

	toRead := stop - m.off
	n, err = m.M.ReadAt(to[:toRead], m.off)
	m.off += int64(n)

	if n < len(to) {
		err = io.EOF
	}

	return
}

func (m *MMapSeeker) Seek(to int64, rel int) (n int64, err error) {
	l := int64(m.M.Len())

	switch rel {
	case io.SeekCurrent:
		m.off += to
	case io.SeekStart:
		m.off = to
	case io.SeekEnd:
		m.off = l + to
	default:
		err = errors.New("invalid whence")
		return
	}

	if m.off < 0 {
		err = errors.New("negative position")
	} else if m.off >= l {
		err = errors.New("past end of file")
	} else {
		n = m.off
	}

	return
}

func (m *MMapSeeker) Close() (err error) {
	return m.M.Close()
}
