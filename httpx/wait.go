package httpx

import (
	"context"
	"sync"
)

type Wait struct {
	count int
	cond  sync.Cond
	once  sync.Once
}

func (w *Wait) init() {
	w.once.Do(func() { w.cond.L = &sync.Mutex{} })
}

func (w *Wait) Add(delta int) {
	w.init()
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	w.count += delta
}

func (w *Wait) Done() {
	w.init()
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	w.count -= 1
}

func (w *Wait) Set(count int) {
	w.init()
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	w.count = count
}

func (w *Wait) Wait(ctx context.Context) {
	w.init()

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
		case <-done:
		}
		w.Set(0)
	}()

	w.cond.L.Lock()
	for w.count > 0 {
		w.cond.Wait()
	}
	w.cond.L.Unlock()
}
