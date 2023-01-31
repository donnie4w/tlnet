package tlnet

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"time"
)

func Data() string {
	return time.Now().Format("2006-01-02")
}

func DataTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func TimeUnix() int64 {
	return time.Now().UnixNano()
}

func Md5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func Octal2bytes(row int32) (bs []byte) {
	bs = make([]byte, 0)
	for i := 0; i < 4; i++ {
		r := row >> uint((3-i)*4)
		bs = append(bs, byte(r))
	}
	return
}

func Bytes2Octal(bb []byte) (value int32) {
	value = int32(0x0000)
	for i, b := range bb {
		ii := uint(b) << uint((3-i)*4)
		value = value | int32(ii)
	}
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
