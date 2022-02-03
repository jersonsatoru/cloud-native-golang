package debounce

import (
	"context"
	"io"
	"net/http"
	"time"
)

var apiCallDebounce Circuit

func init() {
	apiCallDebounce = DebounceFirst(APICall, time.Duration(time.Second*3))
}

func main() {
	http.HandleFunc("/", handleAPI)
	http.ListenAndServe(":8008", nil)
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	res, err := apiCallDebounce(ctx)
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
	request, err := http.NewRequestWithContext(dctx, http.MethodGet, "http://localhost:8009", nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(request)
	if err != nil {
		return "", err
	}
	content, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}
	return string(content), nil
}
