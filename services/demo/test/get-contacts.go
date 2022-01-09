package main

import (
	"encoding/json"
	"github.com/cryptopunkscc/astral-services/components/rpc"
	"github.com/cryptopunkscc/astral-services/services/demo"
	"io"
	"log"
)

func testGetContacts() {
	serverReader, clientWriter := io.Pipe()
	clientReader, serverWriter := io.Pipe()

	clientStream := rwc{
		Reader:      clientReader,
		WriteCloser: clientWriter,
	}
	serverStream := rwc{
		Reader:      serverReader,
		WriteCloser: serverWriter,
	}

	enc := json.NewEncoder(clientStream)
	dec := json.NewDecoder(clientStream)

	go func() {
		enc.Encode(rpc.JsonClientRequest{
			Method: "Messenger.GetContacts",
			Params: nil,
			Id:     0,
		})
	}()

	go func() {
		rpc.ServeCodec(rpc.NewJsonServerCodec(serverStream), demo.Messenger{})
	}()

	response := new(rpc.JsonClientResponse)
	var err error
	for true {
		err = dec.Decode(response)
		if err != nil {
			break
		}
		bytes, _ := json.Marshal(response)
		log.Println(string(bytes))
	}
}
