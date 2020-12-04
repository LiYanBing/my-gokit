package grpc_tool

import (
	"os"
	"os/signal"
	"syscall"
)

func Graceful(do func()) {
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		select {
		case <-ch:
			signal.Stop(ch)
			do()
		}
	}()
}
