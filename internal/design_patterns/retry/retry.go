package retry

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Effector func(ctx context.Context) (string, error)

func Retry(effector Effector, retries int, delay time.Duration) Effector {
	return func(ctx context.Context) (string, error) {
		for r := 0; ; r++ {
			res, err := effector(ctx)
			if err == nil {
				log.Printf("attempt %d worked: %v", r, res)
				return res, err
			}
			if r >= retries {
				return "", fmt.Errorf("retries number exceeded: %d", retries)
			}

			log.Printf("attempt %d failed: retrying in %v", r+1, delay)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}
}
