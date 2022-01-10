package main

import (
	"github.com/cryptopunkscc/astral-services/components/rpc"
	"github.com/cryptopunkscc/astral-services/services/demo"
)

func main() {
	rpc.GenerateSchemaSource("services/demo", demo.ServiceHandle, demo.NewMessenger())
}
