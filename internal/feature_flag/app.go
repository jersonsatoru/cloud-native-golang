package feature_flag

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

var privateCIDRs []*net.IPNet

func init() {
	for _, cidr := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateCIDRs = append(privateCIDRs, block)
	}
}

var enabledFunction map[string]Enabled

func init() {
	enabledFunction = make(map[string]Enabled)
	enabledFunction["use-new-storage"] = fromPrivateIP
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{key}", keyValueGetHandler)
	log.Fatal(http.ListenAndServe(":8009", r))
}

var store map[string]string

type Enabled func(string, *http.Request) (bool, error)

func fromPrivateIP(flag string, r *http.Request) (bool, error) {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false, err
	}
	ip := net.ParseIP(remoteIP)
	if ip == nil {
		return false, errors.New("couldn't parse IP")
	}
	if ip.IsLoopback() {
		return true, nil
	}
	for _, block := range privateCIDRs {
		if block.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}

func keyValueGet(key string) string {
	return store[key]
}

func newKeyValueGet(key string) string {
	log.Println("New key value get")
	return store[key]
}

func keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	var value string
	if featureEnabled("USE_NEW_STORAGE", r) {
		value = newKeyValueGet(key)
	} else {
		value = keyValueGet(key)
	}
	w.Header().Set("content-type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"value": value,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func featureEnabled(flag string, r *http.Request) bool {
	if viper.IsSet(flag) {
		return viper.GetBool(flag)
	}

	enabledFunc, exists := enabledFunction[flag]
	if !exists {
		return false
	}

	result, err := enabledFunc(flag, r)
	if err != nil {
		log.Println(err)
		return false
	}

	return result
}
