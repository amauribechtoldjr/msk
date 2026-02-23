package wipe

import (
	"runtime"
)

func Bytes(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}

	runtime.KeepAlive(buf)
}
