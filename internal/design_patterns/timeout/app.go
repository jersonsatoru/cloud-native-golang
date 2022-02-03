package timeout

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", handleAPI)
	http.ListenAndServe(":8008", nil)
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	apiCallWithTimeout := Timeout(APICall)
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*7)
	defer cancel()
	res, err := apiCallWithTimeout(ctx, "http://localhost:8009/slow")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(res))
}

func APICall(serviceURL string) (string, error) {
	client := http.Client{}
	request, err := http.NewRequest(http.MethodGet, serviceURL, nil)
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
