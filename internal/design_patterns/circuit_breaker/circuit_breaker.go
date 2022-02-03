package circuit_breaker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Circuit func(ctx context.Context) (string, error)

func _() {
	client := &http.Client{
		Timeout: 8 * time.Second,
	}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://www.google.com", nil)
	resp, _ := client.Do(req)
	content, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(content))
}

func Breaker(circuit Circuit, failureThreshold uint) Circuit {
	consecutiveFailures := 0
	lastAttempt := time.Now()
	var m sync.RWMutex
	return func(ctx context.Context) (string, error) {
		m.RLock()
		d := consecutiveFailures - int(failureThreshold)

		if d >= 0 {
			shouldRetryAt := lastAttempt.Add(time.Second * 2 << d)
			if !lastAttempt.After(shouldRetryAt) {
				m.RUnlock()
				return "", errors.New("circuit is open")
			}
		}
		m.RUnlock()
		res, err := circuit(ctx)
		m.Lock()
		defer m.Unlock()
		if err != nil {
			consecutiveFailures++
			return "", err
		}

		consecutiveFailures = 0
		lastAttempt = time.Now()
		return res, err
	}
}
