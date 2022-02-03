package throttle

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	serviceURL = "http://localhost:8009/"
)

func main() {
	http.HandleFunc("/", handleAPI)
	http.ListenAndServe(":8008", nil)
}

var retriableApiCall = Throttle(APICall, 3, 1, time.Duration(5*time.Second))

func handleAPI(w http.ResponseWriter, r *http.Request) {
	res, err := retriableApiCall(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(res))
}

func APICall(ctx context.Context) (string, error) {
	dctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	client := http.Client{}
	request, err := http.NewRequestWithContext(dctx, http.MethodGet, serviceURL, nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(request)
	if err != nil {
		return "", err
	}
	if res.StatusCode > 299 {
		return "", fmt.Errorf("service %s returning internal server error", serviceURL)
	}
	content, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}
	return string(content), nil
}
