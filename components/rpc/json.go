// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"sync"
)

// NewJsonServerCodec returns a new rpc.ServerCodec using JSON-RPC on conn.
func NewJsonServerCodec(conn io.ReadWriteCloser) ServerCodec {
	return &jsonServerCodec{
		dec:     json.NewDecoder(conn),
		enc:     json.NewEncoder(conn),
		c:       conn,
		pending: make(map[uint64]*json.RawMessage),
	}
}

type jsonServerCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.ReadCloser

	// temporary work space
	req JsonServerRequest

	// JSON-RPC clients can use arbitrary json values as request IDs.
	// Package rpc expects uint64 request IDs.
	// We assign uint64 sequence numbers to incoming requests
	// but save the original request ID in the pending map.
	// When rpc responds, we use the sequence number in
	// the response to find the original request ID.
	mutex   sync.Mutex // protects seq, pending
	seq     uint64
	pending map[uint64]*json.RawMessage
}

type IsDone struct {
	IsDone bool
}

type JsonClientRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	Id     uint64        `json:"id"`
}

type JsonClientResponse struct {
	Id     uint64           `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
}

type JsonServerRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	Id     *json.RawMessage `json:"id"`
}

func (r *JsonServerRequest) reset() {
	r.Method = ""
	r.Params = nil
	r.Id = nil
}

type JsonServerResponse struct {
	Id     *json.RawMessage `json:"id"`
	Result interface{}      `json:"result"`
	Error  interface{}      `json:"error"`
}

func (c *jsonServerCodec) ReadRequestHeader(r *Request) error {
	c.req.reset()
	if err := c.dec.Decode(&c.req); err != nil {
		return err
	}
	r.ServiceMethod = c.req.Method

	// JSON request id can be any JSON value;
	// RPC package expects uint64.  Translate to
	// internal uint64 and save JSON on the side.
	c.mutex.Lock()
	c.seq++
	c.pending[c.seq] = c.req.Id
	c.req.Id = nil
	r.Seq = c.seq
	c.mutex.Unlock()

	return nil
}

func (c *jsonServerCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		return nil
	}
	if c.req.Params == nil {
		//return errMissingParams
		return nil
	}
	return json.Unmarshal(*c.req.Params, x)
}

var null = json.RawMessage([]byte("null"))

func (c *jsonServerCodec) WriteResponse(r *Response, x interface{}) error {
	c.mutex.Lock()
	b, ok := c.pending[r.Seq]
	if !ok {
		c.mutex.Unlock()
		return errors.New("invalid sequence number in response")
	}
	//delete(c.pending, r.Seq)
	c.mutex.Unlock()

	if b == nil {
		// Invalid request so no id. Use JSON null.
		b = &null
	}
	resp := JsonServerResponse{Id: b}
	if r.Error == "" {
		resp.Result = x
	} else {
		resp.Error = r.Error
	}
	return c.enc.Encode(resp)
}

func (c *jsonServerCodec) Done() {
	buff := make([]byte, 1)
	reader := c.c
	_, err := reader.Read(buff)
	if err != nil {
		log.Println(err)
		return
	}
}

func (c *jsonServerCodec) Close() error {
	return c.c.Close()
}
