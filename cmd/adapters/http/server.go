package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jersonsatoru/cnb/internal/core"
)

type Runnable interface {
	Start(port string) error
}

type HttpServer struct {
	keyValueStore *core.KeyValueStore
}

func NewHttpServer(keyValueStore *core.KeyValueStore) *HttpServer {
	return &HttpServer{
		keyValueStore: keyValueStore,
	}
}

func (srv *HttpServer) Start(port string) error {
	r := mux.NewRouter()
	r.HandleFunc("/v1/key/{key}", srv.GetHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/key/{key}", srv.PutHandler).Methods(http.MethodPut)
	r.HandleFunc("/v1/key/{key}", srv.PutHandler).Methods(http.MethodDelete)
	server := &http.Server{
		Addr:    port,
		Handler: r,
	}
	return server.ListenAndServeTLS("./cert.pem", "./key.pem")
}

func (srv *HttpServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	value, err := srv.keyValueStore.Get(key)
	if errors.Is(err, core.ErrNoSUchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(value))
}

func (srv *HttpServer) PutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	content, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = srv.keyValueStore.Put(key, string(content))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	location := fmt.Sprintf("/v1/key/%s", key)
	w.Header().Set("location", location)
	w.WriteHeader(http.StatusCreated)
}

func (srv *HttpServer) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	err := srv.keyValueStore.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
