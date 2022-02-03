package resilience

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var GetHostnameThrottled Throttled

func init() {
	GetHostnameThrottled = Throttle(GetHostname, 5, 2, time.Second*3)
}

func GetHostnameHandler(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()
	s := strings.Split(r.RemoteAddr, ":")
	t, str, err := GetHostnameThrottled(ctx, s[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !t {
		http.Error(w, errors.New("too many requests").Error(), http.StatusTooManyRequests)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(str))
}

func GetHostname(ctx context.Context) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return os.Hostname()
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/v1/hostname", GetHostnameHandler)
	log.Fatalln(http.ListenAndServe(":8009", r))
}
