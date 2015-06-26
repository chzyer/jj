package mq

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type MockSub struct {
	size int32
	ch   chan *Msg
	wg   sync.WaitGroup
	Rece []*Msg
}

func NewMockSub() *MockSub {
	m := &MockSub{ch: make(chan *Msg)}
	go func() {
		for {
			msg := <-m.ch
			m.Rece = append(m.Rece, msg)
			m.wg.Done()
			atomic.AddInt32(&m.size, 1)
		}
	}()
	return m
}

func (m *MockSub) Add(n int) {
	m.Rece = m.Rece[:0]
	m.wg.Add(n)
}

func (m *MockSub) Wait() {
	m.wg.Wait()
}

func (m *MockSub) Name() string {
	return "mockSub"
}

func (m *MockSub) ToSelectCase() *reflect.SelectCase {
	return &reflect.SelectCase{
		Dir:  reflect.SelectSend,
		Chan: reflect.ValueOf(m.ch),
	}
}

func TestTopic(t *testing.T) {
	send := [][]byte{[]byte("hello"), []byte("22"), []byte("43333")}
	ms := NewMockSub()
	if true {
		ms.Add(len(send) * 2)
		topic := NewTopic("topic")
		topic.AddSubscriber("ios", ms)
		topic.AddSubscriber("android", ms)
		for _, s := range send {
			topic.Publish(s)
		}
		time.Sleep(time.Millisecond)
		ms.Wait()
		if len(ms.Rece) != len(send)*2 {
			t.Fatal("length of result not except")
		}
		topic.RemoveSubscriber("ios", ms)
		ms.Add(len(send))
		for _, s := range send {
			topic.Publish(s)
		}
		ms.Wait()
		for idx, s := range ms.Rece {
			if !reflect.DeepEqual(send[idx], s.Data) {
				t.Fatal("result not except")
			}
		}
	}

	{
		topic := NewTopic("t2")
		for _, s := range send {
			topic.Publish(s)
		}
		time.Sleep(time.Millisecond)
		ms.Add(len(send))
		topic.AddSubscriber("bb", ms)
		ms.Wait()
		for idx, s := range ms.Rece {
			if !reflect.DeepEqual(send[idx], s.Data) {
				t.Fatal("result not except")
			}
		}
	}
}

func TestChannel(t *testing.T) {
	send := [][]byte{[]byte("hello"), []byte("22"), []byte("43333")}
	ms := NewMockSub()
	ms.Add(len(send))
	ch := NewChannel("topic", "ios")
	ch.AddSubscriber(ms)
	for _, s := range send {
		ch.Write(s)
	}
	ms.Wait()
	for idx, r := range ms.Rece {
		if r.Topic != "topic" || r.channel != "ios" {
			t.Fatal("result not except")
		}
		if !reflect.DeepEqual(r.Data, send[idx]) {
			t.Fatal("result not except")
		}
	}
	ch.RemoveSubscriber(ms)
	for _, s := range send {
		ch.Write(s)
	}
	ms.Add(len(send))
	time.Sleep(time.Millisecond)
	ch.AddSubscriber(ms)
	ms.Wait()
	for idx, r := range ms.Rece {
		if !reflect.DeepEqual(r.Data, send[idx]) {
			t.Fatal("result not except")
		}
	}
}

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
