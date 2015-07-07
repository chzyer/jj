package mq

import (
	"reflect"

	"github.com/jj-io/jj/internal"

	"gopkg.in/logex.v1"
)

import "sync"

type Subscriber interface {
	Name() string
	ToSelectCase() *reflect.SelectCase
}

type Topic struct {
	Name        string
	buffer      chan []byte
	hasChan     chan struct{}
	rat         *internal.Rat
	Chans       []*Channel
	ChanSelect  []reflect.SelectCase
	EmptyChan   reflect.Value
	defaultCase *reflect.SelectCase
	guard       sync.RWMutex
}

func NewTopic(name string, rat *internal.Rat) *Topic {
	var empty chan []byte
	t := &Topic{
		Name:      name,
		rat:       rat,
		hasChan:   make(chan struct{}),
		buffer:    make(chan []byte, 10),
		EmptyChan: reflect.ValueOf(empty),
	}
	t.defaultCase = &reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(t.hasChan),
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
	t.rat.Birth()
	defer t.rat.Die()

	for {
		select {
		case data = <-t.buffer:
			t.writeToChans(data)
		case <-t.rat.C:
			return
		}
	}
}

func (t *Topic) writeToChans(data []byte) {
reWrite:
	vdata := reflect.ValueOf(data)
	t.guard.RLock()
	length := len(t.ChanSelect)
	css := make([]reflect.SelectCase, length+1)
	css[0] = *t.defaultCase
	for idx, cs := range t.ChanSelect {
		cs.Send = vdata
		css[idx+1] = cs
	}
	t.guard.RUnlock()

	hasDefault := true
	for i := 0; i < length || length == 0; i++ {
		chosen, _, _ := reflect.Select(css)
		if chosen == 0 && hasDefault {
			logex.Debug("new subscribe break write op, rewrite")
			goto reWrite
		}
		css[chosen].Chan = t.EmptyChan
		if hasDefault {
			css = css[1:]
			hasDefault = false
		}
	}
	logex.Debugf("topic write to all(%v) chans success", length)
}

func (t *Topic) chanIdx(name string) int {
	for i := 0; i < len(t.Chans); i++ {
		if t.Chans[i].Name == name {
			return i
		}
	}
	return -1
}

func (t *Topic) GetChan(name string) (ch *Channel) {
	t.guard.RLock()
	idx := t.chanIdx(name)
	if idx >= 0 {
		ch = t.Chans[idx]
		t.guard.RUnlock()
		return ch
	}
	t.guard.RUnlock()

	ch = NewChannel(t.Name, name, t.rat)
	t.guard.Lock()
	t.Chans = append(t.Chans, ch)
	t.ChanSelect = append(t.ChanSelect, reflect.SelectCase{
		Dir:  reflect.SelectSend,
		Chan: reflect.ValueOf(ch.underlay),
	})
	t.guard.Unlock()

	select {
	case t.hasChan <- struct{}{}:
	default:
	}
	return
}

func (t *Topic) Publish(data []byte) {
	select {
	case t.buffer <- data:
	case <-t.rat.C:
	}
}
