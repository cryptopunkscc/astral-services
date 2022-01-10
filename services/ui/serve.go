package ui

import (
	"fmt"
	"github.com/cryptopunkscc/astral-services/components/util"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
	"log"
)

const serviceHandle = "ui"

type cmdFunc func(port *astral.Port) error
type cmdMap map[string]cmdFunc

var commands cmdMap
var schemas = make(map[string][]byte)
var schemaSubscribers = make(map[io.ReadWriteCloser]interface{})
var schemaEvents = make(chan interface{})

func init() {
	commands = cmdMap{
		"register":  registerSchema,
		"schema":    streamSchemas,
		"request":   handleRequest,
		"subscribe": handleSubscribe,
	}
}

type portSchema struct {
	port   string
	schema []byte
}

func Serve() {
	astral.Instance().UseTCP = true
	port, err := astral.Reqister(serviceHandle)
	if err != nil {
		log.Panic(err)
	}
	defer port.Close()

	go func() {
		for event := range schemaEvents {
			switch event.(type) {
			case io.ReadWriteCloser:
				listener := event.(io.ReadWriteCloser)
				schemaSubscribers[listener] = listener
				for _, schema := range schemas {
					err := util.WriteL64Bytes(listener, schema)
					if err != nil {
						log.Println(err)
						listener.Close()
						schemaSubscribers[listener] = nil
						break
					}
				}
				log.Println("added listener", listener)
			case portSchema:
				rs := event.(portSchema)
				log.Println("new schema", rs.port)
				schemas[rs.port] = rs.schema
				for subscriber := range schemaSubscribers {
					err := util.WriteL64Bytes(subscriber, rs.schema)
					if err != nil {
						log.Println(err)
						subscriber.Close()
						schemaSubscribers[subscriber] = nil
					}
				}
			}
		}
	}()

	func() {
		queryCounter := 0
		for query := range port.Next() {
			conn, err := query.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			queryCounter++
			go func(conn io.ReadWriteCloser, queryId int) {
				log.Println("reading query")
				queryText, err := enc.ReadL8String(conn)
				if err != nil {
					log.Println("cannot read query")
					return
				}
				log.Println("read query", queryText)
				handle := commands[queryText]
				if handle == nil {
					log.Println("no handle for query ", queryText)
					conn.Close()
					return
				}

				queryHandle := fmt.Sprint("ui-", queryId)
				queryPort, err := astral.Reqister(queryHandle)
				if err != nil {
					log.Println("Cannot register port", queryPort, err)
					return
				}
				defer queryPort.Close()
				enc.WriteL8String(conn, queryHandle)
				handle(queryPort)
			}(conn, queryCounter)
		}
	}()
}

func registerSchema(port *astral.Port) error {
	request, err := (<-port.Next()).Accept()
	if err != nil {
		return err
	}
	serviceHandle, err := enc.ReadL8String(request)
	if err != nil {
		return err
	}
	_, err = request.Write([]byte{})
	if err != nil {
		return err
	}
	schema, err := util.ReadL64Bytes(request)
	if err != nil {
		return err
	}
	schemaEvents <- portSchema{
		port:   serviceHandle,
		schema: schema,
	}
	_, err = request.Write([]byte{})
	return err
}

func streamSchemas(port *astral.Port) error {
	query := <-port.Next()
	conn, err := query.Accept()
	if err != nil {
		return err
	}
	schemaEvents <- conn
	return nil
}

func handleRequest(port *astral.Port) error {
	query := <-port.Next()
	conn, err := query.Accept()
	if err != nil {
		return err
	}
	serviceHandle, err := enc.ReadL8String(conn)
	if err != nil {
		return err
	}
	service, err := astral.Query("", serviceHandle)
	err = util.Join(conn, service)
	return err
}

func handleSubscribe(port *astral.Port) error {
	query := <-port.Next()
	conn, err := query.Accept()
	if err != nil {
		return err
	}
	serviceHandle, err := enc.ReadL8String(conn)
	if err != nil {
		return err
	}
	service, err := astral.Query("", serviceHandle)
	err = util.Join(conn, service)
	return err
}
