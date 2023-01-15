package tlnet

import (
	"bytes"
	"encoding/gob"
)

func encoder(e any) (by []byte, err error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(e)
	by = buf.Bytes()
	return
}

func decoder(buf []byte, e any) (err error) {
	decoder := gob.NewDecoder(bytes.NewReader(buf))
	err = decoder.Decode(e)
	return
}
