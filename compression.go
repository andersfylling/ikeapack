package serialize

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"io"
	"reflect"
)

type compressionReadWriter struct {
	variable
	handler readWriter
}

func (c *compressionReadWriter) readVariable(r io.Reader, v reflect.Value) error {
	lb := make([]byte, 4)
	if _, err := io.ReadFull(r, lb); err != nil {
		return err
	}

	cl := binary.BigEndian.Uint32(lb)
	cb := make([]byte, cl)
	if _, err := io.ReadFull(r, cb); err != nil {
		return err
	}

	z := flate.NewReader(bytes.NewBuffer(cb))
	defer z.Close()

	return handleVariableReader(z, c.handler, v)
}

func (c *compressionReadWriter) writeVariable(w io.Writer, v reflect.Value) error {
	var b bytes.Buffer

	z, err := flate.NewWriter(&b, flate.BestCompression)
	if err != nil {
		return err
	}

	if err := handleVariableWriter(z, c.handler, v); err != nil {
		return err
	}
	z.Close()

	lb := make([]byte, 4)
	binary.BigEndian.PutUint32(lb, uint32(b.Len()))
	if _, err := w.Write(lb); err != nil {
		return err
	}
	if _, err = w.Write(b.Bytes()); err != nil {
		return err
	}

	return nil
}
