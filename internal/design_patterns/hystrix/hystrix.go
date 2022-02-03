package hystrix

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
)

type GoogleGateway interface {
	GetHome()
}

type GoogleGatewayImpl struct{}

func (g *GoogleGatewayImpl) GetHome() (string, error) {
	res, err := http.Get("http://localhost:8009")
	if err != nil {
		return "", err
	}
	if res.StatusCode > 299 {
		return "", fmt.Errorf("response status: %d", res.StatusCode)
	}
	content, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}
	return string(content), err
}

func Breaker(fn func() (string, error)) (string, error) {
	output := make(chan string, 1)
	errors := hystrix.Go(commandName, func() error {
		res, err := fn()
		if err == nil {
			output <- res
		}
		return err
	}, nil)

	select {
	case out := <-output:
		log.Printf("success: %v", out)
		return out, nil
	case err := <-errors:
		log.Printf("failed: %v", err)
		return "", err
	}
}

const (
	commandName = "unreliable_api"
)

func _() {
	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		Timeout:                7000,
		MaxConcurrentRequests:  100,
		ErrorPercentThreshold:  50,
		RequestVolumeThreshold: 3,
		SleepWindow:            20000,
	})
	http.HandleFunc("/", handleAPI)
	http.ListenAndServe(":8008", nil)
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	g := &GoogleGatewayImpl{}
	res, err := Breaker(g.GetHome)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(res))
}
