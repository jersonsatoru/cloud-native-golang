package fan_in

import "sync"

func FanIn(sources ...chan int) chan int {
	var wg sync.WaitGroup
	wg.Add(len(sources))
	out := make(chan int)
	for _, source := range sources {
		go func(ch chan int) {
			defer wg.Done()
			for c := range ch {
				out <- c
			}
		}(source)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
