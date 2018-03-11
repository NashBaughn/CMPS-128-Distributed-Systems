package mainInstance

import (
	"log"
	"net/http"
	// "fmt"
	"encoding/json"
	"httpLogic"
	"kvsAccess"
	"os"
	"partition"
	"regexp"
	"structs"

	"github.com/gorilla/mux"
	// "strconv"
	"io/ioutil"
	"math"
	"net/url"
	"strings"
	"time"
)

var _KVS *kvsAccess.KeyValStore
var _my_node structs.NodeInfo
var _view []structs.NodeInfo
var _K int

func Start() {
	// create router instance
	router := mux.NewRouter()

	//init instance of our global kvs
	_KVS = kvsAccess.NewKVS()

	// fill global vars with ENV vars
	_K = os.Getenv("K")
	_view = viewToStruct(os.Getenv("VIEW")) // regex logic in partition

	// designate handler funcs for each endpoint
	router.HandleFunc("/kvs", newSet).Methods("POST", "PUT")
	router.HandleFunc("/kvs", newGet).Methods("GET")
	router.HandleFunc("/kvs", newDel).Methods("DELETE")
	router.HandleFunc("/kvs/view_update", viewUpdate).Methods("PUT")
	router.HandleFunc("/repartition", repartitionHandler).Methods("PUT")
	router.HandleFunc("/partition", partitionHandler).Methods("PUT")
	router.HandleFunc("/viewchange", addNode).Methods("PUT")
	router.HandleFunc("/kvs/get_number_of_keys", numKeys).Methods("GET")

	// listen on port 8080
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

// // // // // // // // // // // // // //
// 					 Helper Funcs							 //
// // // // // // // // // // // // // //
// regex vars
var _ip = regexp.MustCompile(`(\d+\.){3}(\d+)`)
var _port = regexp.MustCompile(`\d{4}`)
var _n = int((^uint(0)) >> 1)

// converts VIEW string into []structs.IpPort
func viewToStruct(view string) [][]structs.IpPort {
	my_Ip := _ip.FindString(os.Getenv("ip_port"))
	ips := _ip.FindAllString(view, _n)
	ports := _port.FindAllString(view, _n)
	var View = make([][]structs.NodeInfo, math.Ceil(float(len(ips))/float(_K)))
	part_Id := 0
	for i := 0; i < len(ips); i++ {
		temp := structs.NodeInfo{ips[i], ports[i], part_Id}
		if my_Ip == ips[i] {
			_my_node = temp
		}
		View[part_Id] = append(View[part_Id], temp)
		if len(View[part_Id]) == _K {
			part_id++
		}
	}
	return View
}

// converts ip:port string in structs.IpPort
func ipToStruct(ipPort string) structs.IpPort {
	ip := _ip.FindString(ipPort)
	port := _port.FindString(ipPort)
	return structs.IpPort{ip, port}
}

// checks validity of key against constraints
func keyValid(key string) bool {
	keyLen := len(key)
	if keyLen > 250 || keyLen < 1 {
		return false
	}
	return true
}

// returns index of ip in _view
// needs case for when len(view) == 1 ?
func findViewIndex(ip string) int {
	// log.Print("findViewIndex! ip: "+ip)
	for i := 0; i < len(_view); i++ {
		if _view[i].Ip == ip {
			return i
		}
	}
	return -1
}

// removes an element form _view
func removeView(s int) {
	if (len(_view) - 1) == s {
		_view = _view[:s]
	} else {
		_view = append(_view[:s], _view[s+1:]...)
	}
}

// because a lot of error checking occurs
func errPanic(err error) {
	if err != nil {
		panic(err)
	}
}
func errPanicStr(cond bool, err string) {
	if !cond {
		panic(err)
	}
}

// // // // // // // // // // // // // //
// 				 Endpoint Handlers				   //
// // // // // // // // // // // // // //

// PUT Handler for view updates
// This is all in the process of being changed - Riizus
func viewUpdate(w http.ResponseWriter, r *http.Request) {
	// parse request and get relevant info (key, value, view_update, type, ip_port)
	w.Header().Set("Content-Type", "application/json")
	postForm := httpLogic.ViewUpdateForm(r)
	// notify all nodes
	requestList := httpLogic.NotifyNodes(_ip_port, postForm, _view)
	for _, req := range requestList {
		client := &http.Client{Timeout: time.Second}
		// req.ParseForm()
		log.Print("req: " + req.URL.String())
		// log.Print("Node being deleted: "+req.PostForm["ip_port"][0])
		_, err := client.Do(req)
		errPanic(err)
	}
	// if node_addition
	if postForm.Type == "add" {
		// update view
		_view = append(_view, structs.IpPort{postForm.Ip, postForm.Port})
		sendUpdate(postForm)
	} else {
		// update view
		viewIndex := findViewIndex(postForm.Ip)
		errPanicStr(viewIndex != -1, "ip does not exist!")
		removeView(viewIndex)
	}
	// repartition all keys in kvs
	requestMap := partition.Repartition(findViewIndex(_ip_port.Ip), _view, _KVS)
	requestList = httpLogic.CreatePartitionRequests(requestMap)
	// send requests to nodes for repartitioned keys
	for _, req := range requestList {
		client := &http.Client{Timeout: time.Second}
		_, err := client.Do(req)
		errPanic(err)
	}
	// test prints after view_update
	log.Print("view after view_update: ")
	for _, v := range _view {
		log.Print(v.Ip + ":" + v.Port)
	}
	// response logic
	respBody := structs.PartitionResp{"success"}
	bodyBytes, _ := json.Marshal(respBody)
	w.WriteHeader(200)
	w.Write(bodyBytes)
}

// number of keys Handler
func numKeys(w http.ResponseWriter, r *http.Request) {
	numKeys := structs.NumKeys{_KVS.NumKeys()}
	respBody, err := json.Marshal(numKeys)
	errPanic(err)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

func sendUpdate(update structs.ViewUpdateForm) {
	Ip := update.Ip
	Port := update.Port
	URL := "http://" + Ip + ":" + Port + "/viewchange"
	form := url.Values{}
	for _, node := range _view {
		form.Add("Ip", node.Ip)
		form.Add("Port", node.Port)
	}
	formJSON := form.Encode()
	req, _ := http.NewRequest(http.MethodPut, URL, strings.NewReader(formJSON))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	_, err := client.Do(req)
	errPanic(err)
}

func addNode(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	Ip := r.PostForm["Ip"]
	Port := r.PostForm["Port"]
	for i, P := range Ip {
		_view = append(_view, structs.IpPort{P, Port[i]})
	}
	respBody := structs.PartitionResp{"success"}
	bodyBytes, _ := json.Marshal(respBody)
	w.WriteHeader(200)
	w.Write(bodyBytes)

}

func customPrint() {
	log.Print("- - - - - - - - - Custom Print - - - - - - - - -")
	log.Print("          - - - - - Key | Values - - - - -")
	for k, v := range _KVS.Store {
		log.Print("key: " + k + ",   value: " + v)
	}
	log.Print("          - - - - - View - - - - -")
	for _, v := range _view {
		log.Print(v.Ip + ":" + v.Port)
	}
	log.Print("- - - - - - - - - - - - END - - - - - - - - - - -")

}

// Internal endpoint for handling partition notification
func partitionHandler(w http.ResponseWriter, r *http.Request) {
	ipPort := _ip_port.Ip + ":" + _ip_port.Port
	log.Print("partitionHandler: " + ipPort)
	log.Print(ipPort + " view: ")
	for _, v := range _view {
		log.Print(v.Ip + ":" + v.Port)
	}

	// parse request and get relevant info (key, value, view_update, type, ip_port)
	postForm := httpLogic.ViewUpdateForm(r)
	// if node_addition
	if postForm.Type == "add" {
		// update view
		_view = append(_view, structs.IpPort{postForm.Ip, postForm.Port})
	} else {
		// update view
		viewIndex := findViewIndex(postForm.Ip)
		errPanicStr(viewIndex != -1, "ip does not exist!: "+postForm.Ip)
		removeView(viewIndex)
	}
	// repartition all keys in kvs
	requestMap := partition.Repartition(findViewIndex(_ip_port.Ip), _view, _KVS)
	requestList := httpLogic.CreatePartitionRequests(requestMap)
	// send requests to nodes for repartitioned keys
	for _, req := range requestList {
		client := &http.Client{
			Timeout: time.Second,
		}
		_, err := client.Do(req)
		errPanic(err)
	}
	// respond with status
	respBody := structs.PartitionResp{"success"}
	bodyBytes, _ := json.Marshal(respBody)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

// Internal endpoint for handling repartition request body and kvs manipulations
func repartitionHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("repartitionHandler")
	partForm := httpLogic.PartitionForm(r)
	// kvs storage
	for _, v := range partForm {
		log.Print("key: " + v.Key + " value: " + v.Value + " STORED!")
		_KVS.SetValue(v.Key, v.Value)
	}
	// respond with status
	respBody := structs.PartitionResp{"success"}
	bodyBytes, _ := json.Marshal(respBody)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

// new PUT handler for kvs manipulations
func newSet(w http.ResponseWriter, r *http.Request) {
	// var initialization
	put := structs.PUT{0, "yolo", _ip_port.Ip + ":" + _ip_port.Port}
	postForm := httpLogic.PutForm(r)
	key := postForm.Key
	value := postForm.Key
	// check if key exists
	if len(key) == 0 {
		put.Message = "error"
		put.Replaced = 0
		put.Owner = "undetermined"
		w.WriteHeader(400)
		jsonResponse, err := json.Marshal(put)
		errPanic(err)
		w.Write(jsonResponse)
		return
	}
	// key belongs to this node
	if partition.KeyBelongs(key, findViewIndex(_ip_port.Ip), _view) {
		var resp string
		// check validitiy of key
		if !keyValid(key) {
			// repsonse preparation
			put.Message = "error"
			put.Replaced = 0
			put.Owner = "undetermined"
			w.WriteHeader(401)
		} else {
			// do relevant kvs ops
			resp = _KVS.SetValue(key, value)
			// response preparation
			w.Header().Set("Content-Type", "application/json")
			put.Message = "success"
			if resp == "" {
				w.WriteHeader(201)
				put.Replaced = 0
			} else {
				w.WriteHeader(200)
				put.Replaced = 1
			}
		}
		// respond
		jsonResponse, err := json.Marshal(put)
		errPanic(err)
		w.Write(jsonResponse)
		return
	}
	// if key does not belong to node
	// URL logic
	index := partition.KeyBelongsTo(key, _view)
	ipPort := _view[index]
	URL := "http://" + ipPort.Ip + ":" + ipPort.Port + r.URL.Path
	// Request Body Creation
	form := url.Values{}
	form.Add("key", key)
	form.Add("value", value)
	formJSON := form.Encode()
	// Request Creation
	req, err := http.NewRequest(r.Method, URL, strings.NewReader(formJSON))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Request sending logic
	client := &http.Client{
		Timeout: time.Second,
	}
	resp, err := client.Do(req)
	// MainInstance unavailable logic
	if err != nil {
		body := structs.ERROR{"error", "service is not available"}
		jsonBody, _ := json.Marshal(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write(jsonBody)
		return
	}
	// Response handling logic
	defer resp.Body.Close()
	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	errPanic(err2)
	// Response creation logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(bodyBytes)
}

// New GET Handler
func newGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// parse request for relevant data (key, get_number_of_keys)
	key, keyExists := r.URL.Query()["key"]
	var jsonResponse []byte
	var err error
	// check to make sure key exists
	if !keyExists {
		get := structs.ERROR{"no key in request", "keyError"}
		jsonResponse, err = json.Marshal(get)
		errPanic(err)
		w.WriteHeader(400)
		w.Write(jsonResponse)
		return
	}
	// if key belongs to node
	if partition.KeyBelongs(key[0], findViewIndex(_ip_port.Ip), _view) {
		get := structs.GET{"blah", "blah", _ip_port.Ip + ":" + _ip_port.Port}
		if !keyValid(key[0]) {
			w.WriteHeader(401)
			getError := structs.ERROR{"key is empty", "keyError"}
			jsonResponse, err = json.Marshal(getError)
		}
		resp := _KVS.GetValue(key[0])
		if resp == "" {
			w.WriteHeader(404)
			getError := structs.ERROR{"error", "key does not exist"}
			jsonResponse, err = json.Marshal(getError)
		} else {
			w.WriteHeader(200)
			get.Message = "success"
			get.Value = resp
			jsonResponse, err = json.Marshal(get)
		}
		// response logic
		errPanic(err)
		w.Write(jsonResponse)
		return
	}
	// if key does not belong to node
	// URL logic
	index := partition.KeyBelongsTo(key[0], _view)
	ipPort := _view[index]
	URL := "http://" + ipPort.Ip + ":" + ipPort.Port + r.URL.RequestURI()
	// Request Body Creation
	form := url.Values{}
	form.Add("key", key[0])
	formJSON := form.Encode()
	// Request Creation
	req, err := http.NewRequest(r.Method, URL, strings.NewReader(formJSON))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Request sending logic
	client := &http.Client{
		Timeout: time.Second,
	}
	resp, err := client.Do(req)
	// MainInstance unavailable logic
	if err != nil {
		body := structs.ERROR{"error", "service is not available"}
		jsonBody, _ := json.Marshal(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write(jsonBody)
		return
	}
	// Response handling logic
	defer resp.Body.Close()
	jsonResponse, err = ioutil.ReadAll(resp.Body)
	errPanic(err)
	// Response logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(jsonResponse)
}

// New DELETE Handler
func newDel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// parse request for relevant data (key)
	key, keyExists := r.URL.Query()["key"]
	var jsonResponse []byte
	var err error
	// check if key exists
	if !keyExists {
		del := structs.ERROR{"no key in request", "keyError"}
		jsonResponse, err = json.Marshal(del)
		errPanic(err)
		w.WriteHeader(400)
		w.Write(jsonResponse)
		return
	}
	// if key belongs to node
	if partition.KeyBelongs(key[0], findViewIndex(_ip_port.Ip), _view) {
		del := structs.DELExists{"success", _ip_port.Ip + ":" + _ip_port.Port}
		// check validity of key
		if !keyValid(key[0]) {
			w.WriteHeader(401)
			getError := structs.ERROR{"key is either too short or long", "keyError"}
			jsonResponse, err = json.Marshal(getError)
		}
		// kvs logic
		resp := _KVS.DelValue(key[0])
		// create response
		if resp == "" {
			w.WriteHeader(404)
			delError := structs.ERROR{"error", "key does not exist"}
			jsonResponse, err = json.Marshal(delError)
		} else {
			w.WriteHeader(200)
			jsonResponse, err = json.Marshal(del)
		}
		// respond to requester
		errPanic(err)
		w.Write(jsonResponse)
		return
	}
	// if key does not belong to node
	// URL logic
	index := partition.KeyBelongsTo(key[0], _view)
	ipPort := _view[index]
	URL := "http://" + ipPort.Ip + ":" + ipPort.Port + r.URL.RequestURI()
	// log.Print("url: "+URL)
	// Request Body Creation
	form := url.Values{}
	form.Add("key", key[0])
	formJSON := form.Encode()
	// Request Creation
	req, err := http.NewRequest(r.Method, URL, strings.NewReader(formJSON))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Request sending logic
	client := &http.Client{
		Timeout: time.Second,
	}
	resp, err := client.Do(req)
	// MainInstance unavailable logic
	if err != nil {
		body := structs.ERROR{"error", "service is not available"}
		jsonBody, _ := json.Marshal(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write(jsonBody)
		return
	}
	// Response handling logic
	defer resp.Body.Close()
	jsonResponse, err = ioutil.ReadAll(resp.Body)
	errPanic(err)
	// Response logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(jsonResponse)
}
