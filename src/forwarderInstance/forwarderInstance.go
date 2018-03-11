package forwarderInstance

import (
	"log"
	"io/ioutil"
	"fmt"
	"time"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"strings"
	"net/url"
)

// Global Vars
var _ip string
type MainInstanceDown struct {
	Message string `json:"msg"`
	Error string `json:"error"`
}

// Main function
func Start(ip string) {
    _ip = ip
    router := mux.NewRouter()

    router.HandleFunc("/kvs", Forward).Methods("GET", "POST", "PUT", "DELETE")

    if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatal(err)
    }
}

// Forwarding Logic
func Forward(w http.ResponseWriter, r *http.Request) {
	// URL and Request body logic
	home := "http://" + _ip + r.URL.String()
	key := r.PostFormValue("key")
	value := r.PostFormValue("value")
  // Request body creation
	form := url.Values{}
	form.Add("key", key)
	form.Add("value", value)
	formJSON := form.Encode()
	// Request creation
	req, err := http.NewRequest(r.Method, home, strings.NewReader(formJSON))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Request sending logic
	client := &http.Client{
		Timeout : time.Second,
	}
	resp, err := client.Do(req)
	// MainInstance unavailable logic
	if err != nil {
		body := MainInstanceDown{"error", "service is not available"}
		jsonBody, _ := json.Marshal(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write(jsonBody)
		return
	}
	// Response handling logic
	defer resp.Body.Close()
	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	if err2 != nil {
		panic(err2)
	}
	// Response creation logic
	w.WriteHeader(resp.StatusCode)
	fmt.Fprintf(w, "%v", bodyString)
}
