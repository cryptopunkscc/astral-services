package main

import (
	"encoding/json"
	"github.com/cryptopunkscc/astral-services/components/rpc"
	"github.com/cryptopunkscc/astral-services/services/demo"
)

func testGenerateSchema() {
	schema, err := rpc.GenerateSchema(demo.ServiceHandle, demo.Messenger{})
	jsonSchema, err := json.MarshalIndent(schema, "", "  ")
	println(err)
	println(string(jsonSchema))

}
