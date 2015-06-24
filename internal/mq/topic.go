package mq

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type Subscriber interface {
	Name() string
	ToSelectCase() reflect.SelectCase
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

func (t *Topic) Publish(data []byte) {
	if len(t.Chans) == 0 {
		t.buffer <- data
		return
	}
	for i := 0; i < len(t.Chans); i++ {
		t.Chans[i].Write(data)
	}
}

type Channel struct {
	Name              string
	Topic             string
	subscriber        []Subscriber
	selectCase        []reflect.SelectCase
	underlay          chan []byte
	newSubscriberChan chan struct{}
	mutex             sync.Mutex
}

func NewChannel(topic, name string) *Channel {
	ch := &Channel{
		Name:              name,
		underlay:          make(chan []byte, 10),
		newSubscriberChan: make(chan struct{}),
	}
	ch.selectCase = []reflect.SelectCase{
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch.newSubscriberChan)},
	}

	go ch.dispatchLoop()
	return ch
}

func (ch *Channel) dispatchLoop() {
	var (
		data     []byte
		selected int
	)
	for {
		select {
		case data = <-ch.underlay:
		}
		val := reflect.ValueOf(&Msg{
			Topic:   ch.Topic,
			channel: ch.Name,
			Data:    data,
		})

		try := 0
	reDispatch:
		ch.mutex.Lock()
		if len(ch.selectCase) > 0 {
			for i := 1; i < len(ch.selectCase); i++ {
				ch.selectCase[i].Send = val
			}
			selected, _, _ = reflect.Select(ch.selectCase)
		}
		ch.mutex.Unlock()
		//println(ch, "select", selected, len(ch.selectCase))
		if selected == 0 {
			try++
			goto reDispatch
		}
	}
}

func (ch *Channel) AddSubscriber(s Subscriber) {
	ch.mutex.Lock()
	ch.subscriber = append(ch.subscriber, s)
	ch.selectCase = append(ch.selectCase, s.ToSelectCase())
	println(ch, "length:", len(ch.subscriber))
	ch.mutex.Unlock()

	println(2)
	select {
	case ch.newSubscriberChan <- struct{}{}:
	default:
	}
}

func (ch *Channel) RemoveSubscriber(s Subscriber) (idx int) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	idx = -1
	for i := range ch.subscriber {
		if ch.subscriber[i] == s {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}

	ch.subscriber = append(ch.subscriber[:idx], ch.subscriber[idx+1:]...)
	ch.selectCase = append(ch.selectCase[:idx], ch.selectCase[idx+1:]...)
	return
}

func (ch *Channel) Write(data []byte) {
	ch.underlay <- data
}
