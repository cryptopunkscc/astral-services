package demo

import (
	"github.com/cryptopunkscc/astral-services/components/rpc"
	"github.com/cryptopunkscc/astral-services/services/ui"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"log"
)

var ServiceHandle = "demo"

func Serve() {
	astral.Instance().UseTCP = true
	// Prepare messenger
	mess := NewMessenger()

	// Generate scheme
	//asset, err := schema.Asset("schema.json")
	//if err != nil {
	//	log.Panic(err)
	//}
	//var doc *meta_schema.OpenrpcDocument
	//err = json.Unmarshal(asset, doc)
	//if err != nil {
	//	log.Panic(err)
	//}
	//_, err := rpc.GenerateSchema(ServiceHandle, mess)
	//if err != nil {
	//log.Panic(err)
	//}

	done := make(chan error)
	go func() {
		// Register scheme in ui service
		//err := ui.Register(ServiceHandle, doc)
		err := ui.RegisterJson(ServiceHandle, rpc.UnGzip(schema))
		if err != nil {
			done <- err
		}
	}()
	go func() {
		// Serve ui rpc
		done <- rpc.ServeAstral(ServiceHandle, rpc.NewJsonServerCodec, mess)
	}()
	err := <-done
	if err != nil {
		log.Panic(err)
	}
}
