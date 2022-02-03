package debounce

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type Circuit func(context.Context) (string, error)

func DebounceFirst(circuit Circuit, interval time.Duration) Circuit {
	var threshold time.Time
	var result string
	var m sync.Mutex
	return func(ctx context.Context) (string, error) {
		m.Lock()
		log.Printf("%v\n", threshold)
		defer func() {
			threshold = time.Now().Add(interval)
			log.Printf("%v\n", threshold)
			m.Unlock()
		}()
		if time.Now().Before(threshold) {
			return result, errors.New("debounced")
		}
		return circuit(ctx)
	}
}

func DebounceLast(circuit Circuit, interval time.Duration) Circuit {
	var m sync.Mutex
	var once sync.Once
	var threshold time.Time
	var ticker *time.Ticker
	var result string
	var err error

	return func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()
		threshold = time.Now().Add(interval)
		once.Do(func() {
			ticker = time.NewTicker(time.Millisecond * 100)
			go func() {
				defer func() {
					m.Lock()
					ticker.Stop()
					once = sync.Once{}
					m.Unlock()
				}()
			}()

			for {
				select {
				case <-ticker.C:
					m.Lock()
					if time.Now().After(threshold) {
						result, err = circuit(ctx)
						m.Unlock()
						return
					}
					m.Unlock()
				case <-ctx.Done():
					m.Lock()
					result, err = "", ctx.Err()
					m.Unlock()
				}
			}
		})
		return result, err
	}
}
