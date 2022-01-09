package rpc

import (
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
	"log"
)

type GetCodec func(closer io.ReadWriteCloser) ServerCodec

func ServeAstral(
	serviceHandle string,
	getCodec GetCodec,
	rcvrs ...interface{},
) error {
	port, err := astral.Reqister(serviceHandle)
	if err != nil {
		return err
	}

	for req := range port.Next() {
		stream, err := req.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func() {
			err := ServeCodec(getCodec(stream), rcvrs...)
			if err != nil {
				log.Println(err)
			}
		}()
	}
	return nil
}

func ServeCodec(
	codec ServerCodec,
	rcvrs ...interface{},
) (err error) {
	defer func() {
		if err != nil {
			log.Println(err)
		}
		err = codec.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	server := NewServer()

	for _, rcvr := range rcvrs {
		err = server.Register(rcvr)
		if err != nil {
			return
		}
	}

	err = server.ServeRequest(codec)

	return
}
