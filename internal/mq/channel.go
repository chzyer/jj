package mq

import (
	"reflect"
	"sync"

	"github.com/jj-io/jj/internal"
)

type Channel struct {
	Name       string
	Topic      string
	subscriber []Subscriber
	selectCase []*reflect.SelectCase
	underlay   chan []byte
	newChan    chan struct{}
	mutex      sync.Mutex
	rat        *internal.Rat
}

func NewChannel(topic, name string, rat *internal.Rat) *Channel {
	ch := &Channel{
		Name:     name,
		Topic:    topic,
		rat:      rat,
		underlay: make(chan []byte, 10),
		newChan:  make(chan struct{}),
	}
	ch.selectCase = []*reflect.SelectCase{
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch.newChan)},
	}

	go ch.dispatchLoop()
	return ch
}

func (ch *Channel) dispatchLoop() {
	var (
		data     []byte
		selected int
	)
	ch.rat.Birth()
	defer ch.rat.Die()

	for {
		select {
		case data = <-ch.underlay:
		case <-ch.rat.C:
			return
		}
		val := reflect.ValueOf(&Msg{
			Topic:   ch.Topic,
			Channel: ch.Name,
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
	if idx := ch.suberIdx(s); idx >= 0 {
		ch.mutex.Unlock()
		return
	}
	ch.subscriber = append(ch.subscriber, s)
	ch.selectCase = append(ch.selectCase, s.ToSelectCase())
	ch.mutex.Unlock()

	select {
	case ch.newChan <- struct{}{}:
	default:
	}
}

func (ch *Channel) suberIdx(s Subscriber) (idx int) {
	idx = -1
	for i := range ch.subscriber {
		if ch.subscriber[i] == s {
			idx = i
			break
		}
	}
	return
}

func (ch *Channel) RemoveSubscriber(s Subscriber) (idx int) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	idx = ch.suberIdx(s)
	if idx < 0 {
		return
	}

	ch.subscriber = append(ch.subscriber[:idx], ch.subscriber[idx+1:]...)
	ch.selectCase = append(ch.selectCase[:idx+1], ch.selectCase[idx+2:]...)
	return
}

func (ch *Channel) Write(data []byte) {
	select {
	case ch.underlay <- data:
	case <-ch.rat.C:
	}
}
