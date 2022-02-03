package future

import (
	"context"
	"sync"
	"time"
)

type Future interface {
	Result() (string, error)
}

type InnerFuture struct {
	once sync.Once
	wg   sync.WaitGroup

	res   string
	err   error
	resCh <-chan string
	errCh <-chan error
}

func (f *InnerFuture) Result() (string, error) {
	f.once.Do(func() {
		f.wg.Add(1)
		defer f.wg.Done()
		f.res = <-f.resCh
		f.err = <-f.errCh
	})

	f.wg.Wait()
	return f.res, f.err
}

func SlowFunction(ctx context.Context) Future {
	ch := make(chan string)
	err := make(chan error)

	go func() {
		select {
		case <-time.After(time.Second * 2):
			ch <- "i slept for2 seconds"
			err <- nil
		case <-ctx.Done():
			ch <- ""
			err <- ctx.Err()
		}
	}()

	return &InnerFuture{resCh: ch, errCh: err}
}
