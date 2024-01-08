package fastresponse

import (
	"bytes"
	"unsafe"
)

func String2Slice(s string) (b []byte) {
	type StringHeader struct {
		Data uintptr
		Len  int
	}
	type SliceHeader struct {
		Data uintptr
		Len  int
		Cap  int
	}
	bh := (*SliceHeader)(unsafe.Pointer(&b))
	sh := (*StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func Byte2String(b byte) string {
	bs := []byte{b}
	return *(*string)(unsafe.Pointer(&bs))
}

func BytesCombine2(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, String2Slice(""))
}

func ContainsInSlice(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}
