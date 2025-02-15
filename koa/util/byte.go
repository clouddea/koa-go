package util

import (
	"bytes"
	"encoding/binary"
)

func AnyToBytes(n any) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}

func BytesToAny[T any](b []byte) T {
	bytesBuffer := bytes.NewBuffer(b)
	var x T
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}
