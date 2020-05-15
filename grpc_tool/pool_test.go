package grpc_tool

import (
	"testing"
)

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, close, err := Get("127.0.0.1:4096")
		if err != nil {
			b.Errorf("111:%v", err)
			return
		}
		err = close.Close()
		if err != nil {
			b.Errorf("222:%v", err)
		}
	}
}
