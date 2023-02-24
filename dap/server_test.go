package dap

import (
	"testing"
	"time"
)

func TestStartServer(t *testing.T) {
	port := 6666
	svr := StartInstance(port)
	println("Server Started.")

	// 这个过程中，可以执行 client_test

	time.Sleep(20 * time.Second)
	svr.Stop()
	println("Server Stopped")
}
