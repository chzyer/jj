package mq

import (
	"reflect"
	"sync/atomic"
)

type Subscriber interface {
	Name() string
	ToSelectCase() *reflect.SelectCase
}

type Topic struct {
	Name       string
	buffer     chan []byte
	writeState int32
	hasChan    chan struct{}
	Chans      []*Channel
}

func NewTopic(name string) *Topic {
	t := &Topic{
		Name:    name,
		hasChan: make(chan struct{}),
		buffer:  make(chan []byte),
	}
	go t.writeBufferLoop()
	return t
}

func (t *Topic) AddSubscriber(channel string, s Subscriber) {
	t.GetChan(channel).AddSubscriber(s)
}

func (t *Topic) RemoveSubscriber(channel string, s Subscriber) {
	t.GetChan(channel).RemoveSubscriber(s)
}

func (t *Topic) writeBufferLoop() {
	var (
		data []byte
	)
	<-t.hasChan
	for {
		select {
		case data = <-t.buffer:
			t.Publish(data)
		}
	}
}

func (t *Topic) ChanIdx(name string) int {
	for i := 0; i < len(t.Chans); i++ {
		if t.Chans[i].Name == name {
			return i
		}
	}
	return -1
}

func (t *Topic) GetChan(name string) (ch *Channel) {
	if idx := t.ChanIdx(name); idx < 0 {
		ch = NewChannel(t.Name, name)
		t.Chans = append(t.Chans, ch)
		if atomic.CompareAndSwapInt32(&t.writeState, 0, 1) {
			t.hasChan <- struct{}{}
		}
	} else {
		ch = t.Chans[idx]
	}
	return
}

// fixme
func (t *Topic) Publish(data []byte) {
	if len(t.Chans) == 0 {
		t.buffer <- data
		return
	}
	for i := 0; i < len(t.Chans); i++ {
		t.Chans[i].Write(data)
	}
}
