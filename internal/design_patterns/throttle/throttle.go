package throttle

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Effector func(ctx context.Context) (string, error)

func Throttle(e Effector, max uint, refill uint, d time.Duration) Effector {
	var tokens = max
	var once sync.Once
	return func(ctx context.Context) (string, error) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		once.Do(func() {
			ticker := time.NewTicker(d)
			go func() {
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						t := tokens + refill
						if t > max {
							t = max
						}
						log.Printf("Refilling %d", refill)
						tokens = t
					}
				}
			}()
		})
		log.Printf("current tokens: %d", tokens)
		if tokens <= 0 {
			return "", fmt.Errorf("too many calls")
		}
		tokens--
		return e(ctx)
	}
}
