package timeout

import "context"

type SlowFunction func(string) (string, error)

type WithContext func(context.Context, string) (string, error)

func Timeout(slowFunction SlowFunction) WithContext {
	return func(c context.Context, s string) (string, error) {
		chres := make(chan string)
		cherr := make(chan error)
		go func() {
			res, err := slowFunction(s)
			chres <- res
			cherr <- err
		}()

		select {
		case res := <-chres:
			return res, <-cherr
		case <-c.Done():
			return "", c.Err()
		}
	}
}
