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
	// "strconv"
	"io/ioutil"
	"math"
	"net/url"
	"strings"
	"strconv"
	"time"
)

var _KVS *kvsAccess.KeyValStore
var _my_node structs.NodeInfo
var _view [][]structs.NodeInfo
var _K int
var _causal_Payload []int

var testing = false

func Start() {
	//init instance of our global kvs
	_KVS = kvsAccess.NewKVS()

	// fill global vars with ENV vars
	_view = viewToStruct(os.Getenv("VIEW"))
	// _view = viewToStruct(os.Getenv("VIEW")) // regex logic in partition

}

// // // // // // // // // // // // // //
// 					 Helper Funcs							 //
// // // // // // // // // // // // // //
// regex vars
var _ip = regexp.MustCompile(`(\d+\.){3}(\d+)`)
var _port = regexp.MustCompile(`\d{4}`)
var _n = int((^uint(0)) >> 1)

// converts VIEW string into []structs.NodeInfo
// author: Alec
// update: fixed some bugs and added some dummy data
// purpose: the code now runs with no compile or run-time errors/warnings
func viewToStruct(view string) [][]structs.NodeInfo {

	/* Data used in function */

	var my_Ip string
	var _K int
	var ips []string
	var ports []string

	/* Test Data */

	if(testing) {
		my_Ip = _ip.FindString(os.Getenv("10:.0.0.1:8080"))
		_K, _ = strconv.Atoi("3")
		ips = _ip.FindAllString("10.0.0.1:8080, 10.0.0.1:8080, 10.0.0.1:8080, 10.0.0.1:8080", _n)
		ports = _port.FindAllString("10.0.0.1:8080, 10.0.0.1:8080, 10.0.0.1:8080, 10.0.0.1:8080", _n)
	}

	/* Real Data */

	if(!testing) {
		my_Ip = _ip.FindString(os.Getenv("ip_port"))
		_K, _ = strconv.Atoi(os.Getenv("K"))
		ips = _ip.FindAllString(view, _n)
		ports = _port.FindAllString(view, _n)
	}

	/* Print sanity logs */

	log.Print("my_Ip_Port: "+my_Ip)
	// log.Print("ips: "+strings.Join(ips, ""))
	// log.Print("len(ips): "+strconv.Itoa(len(ips)))
	log.Print("_K: "+strconv.Itoa(_K))
	// log.Print("float64(len(ip)): "+strconv.FormatFloat(float64(len(ips)), 'E', -1, 64))
	// log.Print("float64(_K): "+strconv.FormatFloat(float64(_K), 'E', -1, 64))
	// log.Print("float(len(ips)) / float64(_K): "+strconv.FormatFloat(float64(len(ips))/float64(_K), 'E', -1, 64))
	// log.Print(strconv.Itoa(int(math.Ceil(float64(len(ips))/float64(_K)))))

	/* main logic */

	var View = make([][]structs.NodeInfo, int(math.Ceil(float64(len(ips))/float64(_K))))
	part_Id := 0
	for i, ip := range(ips) {
		temp := structs.NodeInfo{ip, ports[i], part_Id, true}
		if my_Ip == ip {
			_my_node = temp
		}
		View[part_Id] = append(View[part_Id], temp)
		if len(View[part_Id]) == _K {
			part_Id++
		}
	}
	log.Print("------------------------------------")
	log.Print("num of partitions: "+strconv.Itoa(len(View)))
	for i, part := range(View) {
		log.Print("partition "+strconv.Itoa(i)+":")
		for k, node := range(part) {
			log.Print(strconv.Itoa(k)+"th Ip: "+node.Ip+" Port: "+node.Port+" Id: "+strconv.Itoa(node.Id)+" Alive: "+strconv.FormatBool(node.Alive))
		}
	}
	log.Print("------------------------------------")
	return View
}

// converts ip:port string in structs.IpPort
// We don't really need this anymore.
func ipToStruct(ipPort string) structs.NodeInfo {
	ip := _ip.FindString(ipPort)
	port := _port.FindString(ipPort)
	return structs.NodeInfo{ip, port, -1, true}
}

// checks validity of key against constraints
func keyValid(key string) bool {
	keyLen := len(key)
	if keyLen > 250 || keyLen < 1 {
		return false
	}
	return true
}

func GetKVS() *kvsAccess.KeyValStore {
	return _KVS
}

func SetKVS(newKVS *kvsAccess.KeyValStore) {
	_KVS = newKVS
}

func GetPayload() []int {
	return _causal_Payload
}

func SetPayload(newPayload []int) {
	_causal_Payload = newPayload
}

// returns index of partition containing
// ip in _view, or -1 if does not exist
func findPartition(ip string) (int, int) {
	// log.Print("findViewIndex! ip: "+ip)
	for i, part := range _view {
		for j, node := range part {
			if (node.Ip == ip) {return i, j}
		}
	}
	return -1, -1
}

// Finds first living node in partition
func findLiving(ind int) structs.NodeInfo {
	var Head structs.NodeInfo
	for _, Head = range _view[ind] {
		if (Head.Alive == true) {
			break
		}
	}
	return Head
}

// removes an element from _view
func removeView(i int, j int) {
	Part := _view[i]
	if (len(Part) - 1) == j {
		Part = Part[:j]
	} else {
		Part = append(Part[:j], Part[j+1:]...)
	}
}

// because a lot of error checking occurs
func ErrPanic(err error) {
	if err != nil {
		panic(err)
	}
}
func ErrPanicStr(cond bool, err string) {
	if !cond {
		panic(err)
	}
}

// // // // // // // // // // // // // //
// 				 Endpoint Handlers				   //
// // // // // // // // // // // // // //

func HBresponse(w http.ResponseWriter, r *http.Request) {
    hb := structs.PartitionResp{"success"}
    jsonResponse, err := json.Marshal(hb)
		ErrPanic(err)
    // maybe include view and/or casual order

    w.WriteHeader(200)
    w.Write(jsonResponse)
}

// PUT Handler for sending view updates
func SendViewUpdate(w http.ResponseWriter, r *http.Request) {
	// parse request and get relevant info (key, value, view_update, type, ip_port)
	postForm := httpLogic.ViewUpdateForm(r)
	// notify all nodes
	requestList := httpLogic.NotifyNodes(_my_node, postForm, _view)
	for _, req := range requestList {
		client := &http.Client{Timeout: time.Second}
		// req.ParseForm()
		log.Print("req: " + req.URL.String())
		// log.Print("Node being deleted: "+req.PostForm["ip_port"][0])
		_, err := client.Do(req)
		ErrPanic(err)
	}
	// if adding new node send updated view table
	if postForm.Type == "add" {
		sendUpdate(postForm)
	}
	// Reuses partition logic
	PartitionHandler(w, r)
}

// number of keys Handler
func NumKeys(w http.ResponseWriter, r *http.Request) {
	numKeys := structs.NumKeys{_KVS.NumKeys()}
	respBody, err := json.Marshal(numKeys)
	ErrPanic(err)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

// Sends all View Info to new node
func sendUpdate(update structs.ViewUpdateForm) {
	Ip := update.Ip
	Port := update.Port
	URL := "http://" + Ip + ":" + Port + "/viewchange"
	form := url.Values{}
	form.Add("my_Ip", Ip)
	form.Add("K", string(_K))
	for _, part := range _view {
		for _, node := range part {
			form.Add("Ip", node.Ip)
			form.Add("Port", node.Port)
			form.Add("Id", string(node.Id))
			form.Add("Alive", strconv.FormatBool(node.Alive))
		}
	}
	formJSON := form.Encode()
	req, _ := http.NewRequest(http.MethodPut, URL, strings.NewReader(formJSON))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	_, err := client.Do(req)
	ErrPanic(err)
}

// Recreates View table and sets node Info
func AddNode(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	my_Ip := r.PostForm["my_Ip"]
	Ip := r.PostForm["Ip"]
	Port := r.PostForm["Port"]
	Id := r.PostForm["Id"]
	Alive := r.PostForm["Alive"]
	_K,_ := strconv.Atoi(r.PostForm["K"][0])

	_view = make([][]structs.NodeInfo, int(math.Ceil(float64(len(Ip))/float64(_K))))
	for i, P := range Ip {
		id, _ := strconv.Atoi(Id[i])
		live, _ := strconv.ParseBool(Alive[i])
		temp := structs.NodeInfo{P, Port[i], id, live}
		if (my_Ip[0] == P) {_my_node = temp}
		_view[id] = append(_view[id], temp)
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
	for _, part := range _view {
		for _, v := range part {
			log.Print(v.Ip + ":" + v.Port)
		}
	}
	log.Print("- - - - - - - - - - - - END - - - - - - - - - - -")

}

// Internal endpoint for handling View Update
func PartitionHandler(w http.ResponseWriter, r *http.Request) {
	// parse request and get relevant info (key, value, view_update, type, ip_port)
	postForm := httpLogic.ViewUpdateForm(r)
	if postForm.Type == "add" {
		// update view
		for i, part := range _view {
			if (len(part) < _K) {
				part = append(part, structs.NodeInfo{postForm.Ip, postForm.Port, i, true})
				break
			}
			if (i+1 == len(_view)) {
				new_part := []structs.NodeInfo{structs.NodeInfo{postForm.Ip, postForm.Port, i+1, true}}
				_view = append(_view, new_part)
				sendRepartition(w)
			}
		}
	} else {
		// update view
		partIndex, nodeIndex := findPartition(postForm.Ip)
		ErrPanicStr(partIndex != -1, "ip does not exist!: "+postForm.Ip)
		removeView(partIndex, nodeIndex)
		if (len(_view[partIndex]) == 0) {
			sendRepartition(w)
		}
	}
}

// repartition all keys in kvs and sends to new partition
func sendRepartition(w http.ResponseWriter) {
	requestMap := partition.Repartition(_my_node.Id, _view, _KVS)
	// Only send requests if the first alive node in partition
	Head := findLiving(_my_node.Id)
	if _my_node == Head {
		requestList := httpLogic.CreatePartitionRequests(_view, requestMap)
		// send requests to nodes for repartitioned keys
		for _, req := range requestList {
			client := &http.Client{
				Timeout: time.Second,
			}
			_, err := client.Do(req)
			ErrPanic(err)
		}
	}
	// respond with status
	respBody := structs.PartitionResp{"success"}
	bodyBytes, _ := json.Marshal(respBody)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

// Internal endpoint for handling repartition request and kvs manipulations
func RepartitionHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("repartitionHandler")
	partForm := httpLogic.PartitionForm(r)
	// kvs storage
	for key, val := range partForm {
		log.Print("key: " + key + " value: " + val + " STORED!")
		_KVS.SetValue(key, val)
	}
	// respond with status
	respBody := structs.PartitionResp{"success"}
	bodyBytes, _ := json.Marshal(respBody)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

// new PUT handler for kvs manipulations
func NewSet(w http.ResponseWriter, r *http.Request) {
	// log.Print("PUT request received")
	// var initialization
	part, _ := findPartition(_my_node.Ip)
	put := structs.NewPUTResp{"success", part, _causal_Payload, time.Now()}
	putForm := httpLogic.PutForm(r)
	log.Print("key: "+putForm.Key)
	log.Print("value: "+putForm.Value)
	// check if key exists
	if len(putForm.Key) == 0 {
		put.Message = "error"
		// put.Replaced = 0
		// put.Owner = "undetermined"
		w.WriteHeader(400)
		jsonResponse, err := json.Marshal(put)
		ErrPanic(err)
		w.Write(jsonResponse)
		return
	}
	// key belongs to this node
	if partition.KeyBelongs(putForm.Key, _my_node.Id, _view) {
		var resp string
		// check validitiy of key
		if !keyValid(putForm.Key) {
			// repsonse preparation
			put.Message = "error"
			// put.Replaced = 0
			// put.Owner = "undetermined"
			w.WriteHeader(401)
		} else {
			// do relevant kvs ops
			resp = _KVS.SetValue(putForm.Key, putForm.Value)
			// response preparation
			w.Header().Set("Content-Type", "application/json")
			put.Message = "success"
			if resp == "" {
				w.WriteHeader(201)
				// put.Replaced = 0
			} else {
				w.WriteHeader(200)
				// put.Replaced = 1
			}
		}
		// respond
		jsonResponse, err := json.Marshal(put)
		ErrPanic(err)
		w.Write(jsonResponse)
		return
	}
	// Not Mine
	genericNotMineResponse(w, r)
}

// New GET Handler
func NewGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// parse request for relevant data (key, get_number_of_keys)
	key, keyExists := r.URL.Query()["key"]
	var jsonResponse []byte
	var err error
	// check to make sure key exists
	if !keyExists {
		get := structs.ERROR{"no key in request", "keyError"}
		jsonResponse, err = json.Marshal(get)
		ErrPanic(err)
		w.WriteHeader(400)
		w.Write(jsonResponse)
		return
	}
	// if key belongs to node
	if partition.KeyBelongs(key[0], _my_node.Id, _view) {
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
			get := structs.GET{"success", resp, _my_node.Ip + ":" + _my_node.Port}
			w.WriteHeader(200)
			jsonResponse, err = json.Marshal(get)
		}
		// response logic
		ErrPanic(err)
		w.Write(jsonResponse)
		return
	}
	// Not Mine
	genericNotMineResponse(w, r)
}

// New DELETE Handler
func NewDel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// parse request for relevant data (key)
	key, keyExists := r.URL.Query()["key"]
	var jsonResponse []byte
	var err error
	// check if key exists
	if !keyExists {
		del := structs.ERROR{"no key in request", "keyError"}
		jsonResponse, err = json.Marshal(del)
		ErrPanic(err)
		w.WriteHeader(400)
		w.Write(jsonResponse)
		return
	}
	// if key belongs to node
	if partition.KeyBelongs(key[0], _my_node.Id, _view) {
		del := structs.DELExists{"success", _my_node.Ip + ":" + _my_node.Port}
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
		ErrPanic(err)
		w.Write(jsonResponse)
		return
	}
	// Not Mine
	genericNotMineResponse(w, r)
}

func genericNotMineResponse(w http.ResponseWriter, r *http.Request) {
	log.Print("---------------------------------")
	log.Print("genericNotMineResponse")
	log.Print("---------------------------------")
	// PostForm logic
	r.ParseForm()
	form := r.PostForm
	// URL logic
	URI := r.URL.RequestURI()
	index := partition.KeyBelongsTo(form["key"][0], _view)
	log.Print("index: "+strconv.Itoa(index))
	ipPort := findLiving(index)
	log.Print("ipPort: "+ipPort.Ip+":"+ipPort.Port)
	URL := "http://" + ipPort.Ip + ":" + ipPort.Port + URI
	log.Print(URL)
	// Request Body Creation
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
	jsonResponse, err := ioutil.ReadAll(resp.Body)
	ErrPanic(err)
	// Response logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(jsonResponse)
}
