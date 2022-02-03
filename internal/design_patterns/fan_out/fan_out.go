package fan_out

func FanOut(source chan int, n int) []chan int {
	dests := make([]chan int, 0)
	for i := 0; i <= n; i++ {
		ch := make(chan int)
		dests = append(dests, ch)
		go func(ch chan int) {
			defer close(ch)
			for s := range source {
				ch <- s
			}
		}(ch)
	}
	return dests
}
