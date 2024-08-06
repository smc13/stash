package stash

import (
	"encoding/json"
	"unsafe"
)

// Implements JSON encoding and BinaryString
// taken from valkey-io/valkey-go
// @see: https://github.com/valkey-io/valkey-go/blob/main/binary.go

func BinaryString(bs []byte) string {
	return unsafe.String(unsafe.SliceData(bs), len(bs))
}

func JSON(in any) string {
	bs, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	return BinaryString(bs)
}
