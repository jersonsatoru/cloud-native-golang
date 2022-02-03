package resilience

import (
	"context"
	"time"
)

type Effector func(context.Context) (string, error)

type Throttled func(ctx context.Context, callerIdentity string) (bool, string, error)

type bucket struct {
	tokens uint
	time   time.Time
}

func Throttle(e Effector, max uint, refill uint, d time.Duration) Throttled {
	buckets := make(map[string]*bucket)
	return func(ctx context.Context, callerIdentity string) (bool, string, error) {
		b := buckets[callerIdentity]
		if b == nil {
			buckets[callerIdentity] = &bucket{tokens: max - 1, time: time.Now()}
			str, err := e(ctx)
			return true, str, err
		}

		refillInterval := uint(time.Since(b.time) / d)
		tokensAdded := refill * refillInterval
		currentTokens := b.tokens + tokensAdded

		if currentTokens < 1 {
			return false, "", nil
		}

		if currentTokens > max {
			b.time = time.Now()
			b.tokens = max - 1
		} else {
			deltaTokens := currentTokens - b.tokens
			deltaRefills := deltaTokens / refill
			deltaTime := time.Duration(deltaRefills) * d

			b.time = b.time.Add(deltaTime)
			b.tokens = currentTokens - 1
		}

		str, err := e(ctx)
		return true, str, err
	}
}
