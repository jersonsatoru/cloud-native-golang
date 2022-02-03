package fan_out

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	source := make(chan int)
	dests := FanOut(source, 3)

	go func() {
		for i := 1; i <= 10; i++ {
			time.Sleep(time.Millisecond * 500)
			source <- i
		}
		close(source)
	}()

	var wg sync.WaitGroup
	wg.Add(len(dests))

	for i, ch := range dests {
		go func(ch chan int, i int) {
			defer wg.Done()
			for c := range ch {
				fmt.Printf("channel %d, value %d\n", i, c)
			}
		}(ch, i)
	}

	wg.Wait()
}
