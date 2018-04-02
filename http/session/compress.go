package session

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
)

// gzipWrite reads from the slice of bytes and writes the compressed data to the
// writer
func GzipWrite(w io.Writer, data []byte) error {
	// Write gzipped data to the client
	gw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gw.Close()
	gw.Write(data)
	return err
}

// gunzipWrite reads from the gzipped slice of bytes and writes the uncompressed
// data to the writer
func GunzipWrite(w io.Writer, data []byte) error {
	// Write gzipped data to the client
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer gr.Close()
	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}
