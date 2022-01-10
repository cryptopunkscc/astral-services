package rpc

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
)

func UnGzip(data []byte) (bytes []byte) {
	reader, writer := io.Pipe()
	go func() {
		_, err := writer.Write(data)
		if err != nil {
			log.Panic(err)
		}
		_ = writer.Close()
	}()
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		log.Panic(err)
	}
	bytes, err = io.ReadAll(gzipReader)
	if err != nil {
		log.Panic(err)
	}
	return
}

func GenerateSchemaSource(path, name string, rec interface{}) {
	pkgPath := reflect.ValueOf(rec).Elem().Type().PkgPath()
	_, pkg := filepath.Split(pkgPath)
	doc, err := GenerateSchema(name, rec)
	if err != nil {
		log.Panic(err)
	}
	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		log.Panic(err)
	}
	println(jsonDoc)

	dir, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}
	filePath := filepath.Join(dir, path, "schema.go")
	file, err := os.Create(filepath.Join(filePath))
	if err != nil {
		log.Panic(err)
	}
	_, _ = file.Write([]byte("package " + pkg + "\n\n"))
	err = writeFiles(file, jsonDoc, "schema")
	if err != nil {
		log.Panic(err)
	}

	_ = file.Close()
}

const lowerHex = "0123456789abcdef"

type StringWriter struct {
	io.Writer
	c int
}

func (w *StringWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}
	buf := []byte(`\x00`)
	var b byte
	for n, b = range p {
		buf[2] = lowerHex[b/16]
		buf[3] = lowerHex[b%16]
		w.Writer.Write(buf)
		w.c++
	}
	n++
	return
}

func writeFiles(w io.Writer, scheme []byte, name string) error {
	_, err := fmt.Fprintf(w, `var %s = []byte("`, name)
	if err != nil {
		return err
	}

	gz := gzip.NewWriter(&StringWriter{Writer: w})
	gz.Write(scheme)
	gz.Close()

	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, `")`)

	return nil
}
