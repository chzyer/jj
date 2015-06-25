package mq

import (
	"reflect"
	"sync"
)

type Channel struct {
	Name              string
	Topic             string
	subscriber        []Subscriber
	selectCase        []*reflect.SelectCase
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
	ch.selectCase = []*reflect.SelectCase{
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

	reDispatch:

		ch.mutex.Lock()
		sc := make([]reflect.SelectCase, len(ch.selectCase))
		for i := 0; i < len(ch.selectCase); i++ {
			sc[i] = *ch.selectCase[i]
			if i > 0 {
				sc[i].Send = val
			}
		}
		ch.mutex.Unlock()
		selected, _, _ = reflect.Select(sc)
		if selected == 0 {
			goto reDispatch
		}
	}
}

func (ch *Channel) AddSubscriber(s Subscriber) {
	ch.mutex.Lock()
	ch.subscriber = append(ch.subscriber, s)
	ch.selectCase = append(ch.selectCase, s.ToSelectCase())
	ch.mutex.Unlock()

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
	ch.selectCase = append(ch.selectCase[:idx+1], ch.selectCase[idx+2:]...)
	return
}

func (ch *Channel) Write(data []byte) {
	ch.underlay <- data
}
