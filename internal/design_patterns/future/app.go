package future

import (
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	res := SlowFunction(ctx)
	data, err := res.Result()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(data)
}
