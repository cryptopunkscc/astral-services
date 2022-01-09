package util

import (
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/enc"
	"io"
)

func WriteL64Bytes(w io.Writer, bytes []byte) error {
	var err error
	var l = len(bytes)

	err = enc.Write(w, uint64(l))
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

func ReadL64Bytes(r io.Reader) ([]byte, error) {
	l, err := ReadUint64(r)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, l)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func ReadUint64(r io.Reader) (i uint64, err error) {
	err = binary.Read(r, binary.BigEndian, &i)
	return
}
