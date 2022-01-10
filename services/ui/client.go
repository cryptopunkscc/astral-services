package ui

import (
	"encoding/json"
	"github.com/cryptopunkscc/astral-services/components/util"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	openrpc_document "github.com/open-rpc/meta-schema"
)

func Register(port string, doc *openrpc_document.OpenrpcDocument) (err error) {
	conn, err := astral.Query("", serviceHandle)
	defer conn.Close()
	if err != nil {
		return
	}
	err = enc.WriteL8String(conn, "register")
	if err != nil {
		return
	}
	queryPort, err := enc.ReadL8String(conn)
	if err != nil {
		return
	}
	queryConn, err := astral.Query("", queryPort)
	defer queryConn.Close()
	if err != nil {
		return
	}
	err = enc.WriteL8String(queryConn, port)
	if err != nil {
		return
	}
	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		return
	}
	err = util.WriteL64Bytes(queryConn, jsonDoc)
	return
}

func RegisterJson(port string, doc []byte) (err error) {
	conn, err := astral.Query("", serviceHandle)
	defer conn.Close()
	if err != nil {
		return
	}
	err = enc.WriteL8String(conn, "register")
	if err != nil {
		return
	}
	queryPort, err := enc.ReadL8String(conn)
	if err != nil {
		return
	}
	queryConn, err := astral.Query("", queryPort)
	defer queryConn.Close()
	if err != nil {
		return
	}
	err = enc.WriteL8String(queryConn, port)
	if err != nil {
		return
	}
	err = util.WriteL64Bytes(queryConn, doc)
	return
}
