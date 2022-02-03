package resilience

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type ClientContext struct {
	http.Client
}

func (c *ClientContext) GetContext(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func TimeoutHandler(w http.ResponseWriter, r *http.Request) {
	cc := &ClientContext{}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()
	resp, err := cc.GetContext(ctx, "http://localhost:8008/slow")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(content))
}

func _() {
	r := mux.NewRouter()
	r.HandleFunc("/", TimeoutHandler)
	log.Fatal(http.ListenAndServe(":8009", r))
}
