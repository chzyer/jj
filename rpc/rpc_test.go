package rpc

import "testing"

func TestAll(t *testing.T) {
	handler := &TcpHandler{}
	go func() {
		err := Listen(":12345", handler)
		if err != nil {
			t.Fatal(err)
		}
	}()

}
