package resilience

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const MaxQueueDepth = 1000

func CurrentQueue() int {
	return 1040
}

func loadSheddingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CurrentQueue() > MaxQueueDepth {
			log.Println("load shedding engaged")
			http.Error(w, "load shedding engaged", http.StatusServiceUnavailable)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getHostnameHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hostname"))
}

func _() {
	r := mux.NewRouter()
	r.Handle("/v1/hostname", loadSheddingMiddleware(http.HandlerFunc(getHostnameHandler)))
	log.Fatal(http.ListenAndServe(":8009", r))
}
