package fan_in

import (
	"fmt"
	"time"
)

func main() {
	sources := make([]chan int, 0)
	for i := 0; i <= 3; i++ {
		ch := make(chan int)
		sources = append(sources, ch)
		go func(c chan int, channelNumber int) {
			defer close(c)
			for j := 0; j <= 5; j++ {
				time.Sleep(500 * time.Millisecond)
				c <- j
			}
		}(ch, i)
	}

	out := FanIn(sources...)
	for dest := range out {
		fmt.Println(dest)
	}
}
