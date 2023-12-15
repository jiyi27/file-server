package main

import (
	"io"
)

type LimitedReader struct {
	r io.ReadCloser
	n int64
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.n <= 0 {
		return 0, &FileToLarge{Message: "limited reader: file too large"}
	}
	if len(p) == 0 {
		return 0, nil
	}

	if int64(len(p)) > l.n {
		p = p[0:l.n]
	}
	n, err = l.r.Read(p)
	l.n -= int64(n)
	return n, err
}

func (l *LimitedReader) Close() error {
	return l.r.Close()
}

type FileToLarge struct {
	Message string
}

func (e *FileToLarge) Error() string {
	return e.Message
}
