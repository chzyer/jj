package mq

import (
	"testing"
	"time"
)

func TestMq(t *testing.T) {
	mq := NewMq()
	client := NewMqClient(mq)
	client.Subscribe("to:me", "ios")
	client.Subscribe("to:me", "android")
	mq.Publish("to:me", []byte("suck"))
	for i := 0; i < 2; i++ {
		select {
		case msg := <-client.RespChan:
			if string(msg.Data) != "suck" {
				t.Log(msg)
				t.Error("sub not except")
			}
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

}
