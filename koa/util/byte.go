package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func AnyToBytes(n any) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	Assert(binary.Write(bytesBuffer, binary.BigEndian, n), fmt.Sprintf("invalid value: %v", n))
	return bytesBuffer.Bytes()
}

func BytesToAny[T any](b []byte) T {
	bytesBuffer := bytes.NewBuffer(b)
	var x T
	Assert(binary.Read(bytesBuffer, binary.BigEndian, &x), "can not read bytes")
	return x
}
