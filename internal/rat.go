package internal

import "sync"

type Rat struct {
	C chan struct{}
	W sync.WaitGroup
}

func NewRat() *Rat {
	return &Rat{
		C: make(chan struct{}),
	}
}

func (r *Rat) Shoo() {
	close(r.C)
	r.W.Wait()
}

func (r *Rat) Birth() {
	r.W.Add(1)
}

func (r *Rat) Kill() {
	r.W.Done()
}
