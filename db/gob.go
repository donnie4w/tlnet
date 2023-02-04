package db

import (
	"bytes"
	"encoding/binary"
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

func Int64ToBytes(n int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}

func BytesToInt64(bs []byte) (_r int64) {
	bytesBuffer := bytes.NewBuffer(bs)
	binary.Read(bytesBuffer, binary.BigEndian, &_r)
	return
}
