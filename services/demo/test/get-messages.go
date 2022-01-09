package main

import (
	"encoding/json"
	"github.com/cryptopunkscc/astral-services/components/rpc"
	"github.com/cryptopunkscc/astral-services/services/demo"
	"io"
	"log"
	"time"
)

func testGetMessages() {
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
		req := rpc.JsonClientRequest{
			Method: "Messenger.GetMessages",
			Params: []interface{}{
				demo.ContactId{Value: "1"},
			},
			Id: 0,
		}
		reqBytes, _ := json.MarshalIndent(req, "", "  ")
		log.Println(string(reqBytes))
		enc.Encode(req)
	}()

	go func() {
		rpc.ServeCodec(rpc.NewJsonServerCodec(serverStream), demo.NewMessenger())
	}()

	go func() {
		time.Sleep(time.Second)
		clientStream.Write([]byte{})
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
