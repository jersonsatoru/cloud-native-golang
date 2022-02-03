package hot_reload_config

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", GetConfigHandler)
	log.Fatal(http.ListenAndServe(":8009", r))
}

func GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(ConfigFile)
}
