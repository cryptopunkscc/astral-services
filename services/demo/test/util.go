package main

import "io"

type rwc struct {
	io.Reader
	io.WriteCloser
}

func (r rwc) Close() error {
	return r.WriteCloser.Close()
}
