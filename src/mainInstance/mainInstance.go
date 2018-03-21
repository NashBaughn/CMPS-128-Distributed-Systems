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
	//"heartMonitor"
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
	log.Print("view: "+os.Getenv("VIEW"))
	_view = viewToStruct(os.Getenv("VIEW"))
	PrintView()

	for _, _ = range _view {
		_causal_Payload = append(_causal_Payload, 0)
	}

	log.Print("_my_node:")
	log.Print(_my_node)
	log.Print("My _K: "+strconv.Itoa(_K))
	// start heartMonitor

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
	log.Print("viewToStruct")

	/* Data used in function */
	var my_Ip string
	var ips []string
	var ports []string

	/* Test Data */

	if(testing) {
		my_Ip = _ip.FindString(os.Getenv("10:.0.0.1:8080"))
		_K = 3
		ips = _ip.FindAllString("10.0.0.1:8080, 10.0.0.2:8080, 10.0.0.3:8080, 10.0.0.4:8080", _n)
		ports = _port.FindAllString("10.0.0.1:8080, 10.0.0.1:8080, 10.0.0.1:8080, 10.0.0.1:8080", _n)
	}

	/* Real Data */

	if(!testing) {
		my_Ip = _ip.FindString(os.Getenv("ip_port"))
		log.Print("Real _K:" + os.Getenv("K"))
		log.Print(strconv.Atoi(os.Getenv("K")))
		K, _ := strconv.ParseInt(os.Getenv("K"),0,64)
		_K = int(K)
		log.Print("Temp _K: "+string(_K))
		ips = _ip.FindAllString(view, _n)
		ports = _port.FindAllString(view, _n)
	}

	/* Print sanity logs */
	log.Print("my_Ip_Port: "+my_Ip)
	// log.Print("ips: "+strings.Join(ips, ""))
	// log.Print("len(ips): "+strconv.Itoa(len(ips)))

	// log.Print("float64(len(ip)): "+strconv.FormatFloat(float64(len(ips)), 'E', -1, 64))
	// log.Print("float64(_K): "+strconv.FormatFloat(float64(_K), 'E', -1, 64))
	// log.Print("float(len(ips)) / float64(_K): "+strconv.FormatFloat(float64(len(ips))/float64(_K), 'E', -1, 64))
	// log.Print(strconv.Itoa(int(math.Ceil(float64(len(ips))/float64(_K)))))

	/* main logic */

	// Nash - hey so this allows the program to be run without testing = true
	if (_K == 0) {
		return nil
	}
	log.Print("Ips Length: "+string(len(ips)))
	var View = make([][]structs.NodeInfo, int(math.Ceil(float64(len(ips))/float64(_K))))
	log.Print("View Length: "+string(len(View)))
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

	return View
}

func GetView() [][]structs.NodeInfo {
	return _view
}

func GetNode() structs.NodeInfo {
	return _my_node
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

func StringifyPayload() string {
	var payload_string = ""
	for _, load := range(_causal_Payload) {
		payload_string += (","+strconv.Itoa(load))
	}
	return payload_string
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

// Finds all living nodes in partition
func findAllLiving(node structs.NodeInfo) []structs.NodeInfo {
	var Alive []structs.NodeInfo
	for _, Head := range _view[node.Id] {
		if (Head.Alive == true && Head != node) {
			Alive = append(Alive, Head)
		}
	}
	return Alive
}

// removes an element from _view
func removeView(i int, j int) bool{
	log.Print("In Remove View, removing:")
	shift := false
	//var OldNode structs.NodeInfo
	Part := _view[i]
	log.Print(Part[j])
	if (len(Part) - 1) == j {
	 	_view[i] = Part[:j]
	} else {
 		_view[i] = append(Part[:j], Part[j+1:]...)
	}
	if (i != len(_view)-1) {
		OldPart := _view[len(_view)-1]
		//log.Print(OldPart)
		OldNode := OldPart[len(OldPart)-1]
		if (_my_node.Ip == OldNode.Ip) {
			_causal_Payload[_my_node.Id] = 0
			_my_node.Id = i
		} else if (_my_node.Id == i){
			SendKVSMend(OldNode)
		}
		OldNode.Id = i
		_view[i] = append(_view[i], OldNode)
		if (len(OldPart) == 1){
			_view = _view[:len(_view)-1]
			//_causal_Payload = _causal_Payload[:len(_causal_Payload)-1]
			shift = true
		}else {
			_view[len(_view)-1] = OldPart[:len(OldPart)-1]
		}
		//log.Print(_view[len(_view)-1])
	} else {
		if (len(_view[i]) == 0){
			_view = _view[:len(_view)-1]
			//_causal_Payload = _causal_Payload[:len(_causal_Payload)-1]
			shift = true
		}
	}
	return shift
}

func SendKVSMend (Node structs.NodeInfo) {
	log.Print("In Send KVS Mend")
	log.Print("Sending new KVS to: " + Node.Ip)
	if (len(_view[Node.Id]) ==1) {
		log.Print("Waiting for node to repartition")
		time.Sleep(2 * time.Second)
	}
    Ip := Node.Ip
	Port := Node.Port
	URL := "http://" + Ip + ":" + Port + "/KVSMend"
	form := url.Values{}
	form.Add("sender", _my_node.Ip)
	for key, val := range GetKVS().Store {
		form.Add("Key", key)
		form.Add("Val", val)
	}
    for _, num := range _causal_Payload {
        form.Add("Payload", strconv.Itoa(num))
    }
	formJSON := form.Encode()
	req, _ := http.NewRequest(http.MethodPut, URL, strings.NewReader(formJSON))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	_, err := client.Do(req)
    if err != nil {
		panic(err)
	}
}

// Retrieves new KVS from other node in partition
// Checks the Causal Payload to see if it is newer than current one
func HandleKVSMend (w http.ResponseWriter, r *http.Request) {
	log.Print("In Handle KVS Mend")
    r.ParseForm()
	newer := true
    Payload := r.PostForm["Payload"]
	log.Print("requestings host: "+r.FormValue("sender"))
    var newPayload []int
	log.Print("Length of my Payload: "+strconv.Itoa(len(_causal_Payload))+" Length of new Payload: "+strconv.Itoa(len(Payload)))
    for ind, my_num := range _causal_Payload {
		log.Print("Index: "+strconv.Itoa(ind))
        new_num, _  := strconv.Atoi(Payload[ind])
		log.Print("Num: "+Payload[ind])
        if (my_num > new_num) {
			log.Print("Older KVS")
            newer = false
            break
        }
        newPayload = append(newPayload,new_num)
    }
    if newer {
        _KVS = kvsAccess.NewKVS()
		keys := r.PostForm["Key"]
		vals := r.PostForm["Val"]
        for i, key := range keys {
			log.Print("Adding Key: "+key + " Value: " + vals[i])
            _KVS.SetValue(key, vals[i])
        }
        _causal_Payload = newPayload
    }
    respBody := structs.PartitionResp{"success"}
    bodyBytes, _ := json.Marshal(respBody)
    w.WriteHeader(200)
    w.Write(bodyBytes)
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
func PrintView() {
	log.Print("------------------------------------")
	log.Print("num of partitions: "+strconv.Itoa(len(_view)))
	for i, part := range(_view) {
		log.Print("partition "+strconv.Itoa(i)+":")
		for k, node := range(part) {
			log.Print(strconv.Itoa(k)+"th Ip: "+node.Ip+" Port: "+node.Port+" Id: "+strconv.Itoa(node.Id)+" Alive: "+strconv.FormatBool(node.Alive))
		}
	}
	log.Print("------------------------------------")
}

// // // // // // // // // // // // // //
// 				 Endpoint Handlers				   //
// // // // // // // // // // // // // //

func HBresponse(w http.ResponseWriter, r *http.Request) {
    hb := structs.PartitionResp{"success"}
    jsonResponse, err := json.Marshal(hb)
		ErrPanic(err)
    // maybe include view and/or casual order

    //w.WriteHeader(200)
    w.Write(jsonResponse)
}

func HelpMe(w http.ResponseWriter, r *http.Request) {
    ok := structs.PartitionResp{"success"}
    jsonResponse, err := json.Marshal(ok)
	ErrPanic(err)
	Part := _view[_my_node.Id]
	SendKVSMend(Part[len(Part)-1])
    // maybe include view and/or casual order
    //w.WriteHeader(200)
    w.Write(jsonResponse)
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

func GetAllPartitionIds(w http.ResponseWriter, r *http.Request) {
	var partition_ids []int
	for i, _ := range(_view) {
		partition_ids = append(partition_ids, i)
	}
	form := structs.GETAllPartitionIdsResp{"success", partition_ids}
	respBody, err := json.Marshal(form)
	ErrPanic(err)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

func GetPartitionId(w http.ResponseWriter, r *http.Request) {
	form := structs.GETPartitionIdResp{"success", _my_node.Id}
	respBody, err := json.Marshal(form)
	ErrPanic(err)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

func GetPartitionMembers(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var partition_members []string
	id, _ := strconv.Atoi(r.Form["partition_id"][0])
	for _, node := range(_view[id]) {
		partition_members = append(partition_members, node.Ip+":"+node.Port)
	}
	form := structs.GETPartitionMembersResp{"success", partition_members}
	respBody, err := json.Marshal(form)
	ErrPanic(err)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

// PUT Handler for sending view updates
func SendViewUpdate(w http.ResponseWriter, r *http.Request) {
	log.Print("SendViewUpdate")
	// parse request and get relevant info (key, value, view_update, type, ip_port)
	postForm := httpLogic.ViewUpdateForm(r)
	// notify all nodes
	requestList := httpLogic.NotifyNodes(_my_node, postForm, _view)
	for _, req := range requestList {
		client := &http.Client{Timeout: time.Second}
		// req.ParseForm()
		log.Print("req: " + req.URL.String())
		// log.Print("Node being deleted: "+req.PostForm["ip_port"][0])
		_, _ = client.Do(req)
		//ErrPanic(err)
	}
	// Reuses partition logic
	PartitionHandler(w, r)
	// if adding new node send updated view table
	if postForm.Type == "add" {
		sendUpdate(postForm)
	}
}

// Internal endpoint for handling View Update
func PartitionHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("PartitionHandler")
	// parse request and get relevant info (key, value, view_update, type, ip_port)
	log.Print("In Partition Handler")
	postForm := httpLogic.ViewUpdateForm(r)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	if postForm.Type == "add" {
		log.Print("Adding new Node:" + postForm.Ip)
		// update view
		log.Print("view before add: ")
		PrintView()
		var respBody structs.AddNodeResp
		for i, part := range _view {
			if (len(part) < _K) {
				newNode:= structs.NodeInfo{postForm.Ip, postForm.Port, i, true}
				_view[i] = append(part, structs.NodeInfo{postForm.Ip, postForm.Port, i, true})
				if (_my_node.Id == i){
					SendKVSMend(newNode)
				}
				break
			}
			if (i+1 == len(_view)) {
				new_part := []structs.NodeInfo{structs.NodeInfo{postForm.Ip, postForm.Port, i+1, true}}
				_view = append(_view, new_part)
				log.Print("Sending Repartition")
				sendRepartition(_my_node.Id, _view)
				_causal_Payload = append(_causal_Payload,0)
			}
		}
		log.Print("view after add:")
		PrintView()
		// respond with status
		ind,_ :=findPartition(postForm.Ip)
		respBody = structs.AddNodeResp{"success",ind, len(_view)}
		bodyBytes, _ := json.Marshal(respBody)
		w.Write(bodyBytes)
	} else {
		log.Print("Removing Node:" + postForm.Ip)
		// update view
		log.Print("view before remove:")
		PrintView()
		Part := _view[len(_view)-1]
		Temp := Part[len(Part)-1]
		if (_my_node.Ip == Temp.Ip && len(Part) == 1){
			log.Print("Im moving need to Repartition")
			sendRepartition(-1, _view[:len(_view)-1])
		} else {
			log.Print("Waiting for node to repartition")
			time.Sleep(2*time.Second)
		}
		partIndex, nodeIndex := findPartition(postForm.Ip)
		ErrPanicStr(partIndex != -1, "ip does not exist!: "+postForm.Ip)
		if removeView(partIndex, nodeIndex) {
			if (_my_node.Ip != Temp.Ip || len(Part) != 1){

				sendRepartition(_my_node.Id, _view)
			}
			_causal_Payload = _causal_Payload[:len(_causal_Payload)-1]
		}

		/*if (shift && node.Ip == _my_node.Ip){
			for _, friend := range _view[node.Id] {
				http.Get("http://" + friend.Ip + ":" + friend.Port + "/helpMe")
			}
			_my_node.Id = node.Id
		}*/
		log.Print("view after remove:")
		PrintView()
		respBody := structs.RemoveNodeResp{"success", len(_view)}
		bodyBytes, _ := json.Marshal(respBody)
		w.Write(bodyBytes)
	}
	log.Print(_view)
	log.Print("Leaving Partition Handler")
}

// repartition all keys in kvs and sends to new partition
func sendRepartition(ind int, view [][]structs.NodeInfo) {
	log.Print("In Send Repartition")
	log.Print("Num Keys before removing: "+strconv.Itoa(_KVS.NumKeys()))
	requestMap := partition.Repartition(ind, view, _KVS)
	log.Print("Num Keys adter removing: "+strconv.Itoa(_KVS.NumKeys()))
	// Only send requests if the first alive node in partition
	Head := findLiving(_my_node.Id)
	if (_my_node == Head || ind == -1){
		requestList := httpLogic.CreatePartitionRequests(view, requestMap)
		// send requests to nodes for repartitioned keys
		for _, req := range requestList {
			client := &http.Client{
				Timeout: time.Second,
			}
			_, err := client.Do(req)
			ErrPanic(err)
		}
	}
	log.Print("Leaving Send Repartition")
}

// Internal endpoint for handling repartition request and kvs manipulations
func RepartitionHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("In Repartition Handler")
	partForm := httpLogic.PartitionForm(r)
	// kvs storage
	log.Print("Num Keys before adding: "+strconv.Itoa(_KVS.NumKeys()))
	for key, val := range partForm {
		log.Print("key: " + key + " value: " + val + " STORED!")
		_KVS.SetValue(key, val)
	}
	log.Print("Num Keys after adding: "+strconv.Itoa(_KVS.NumKeys()))
	// respond with status
	respBody := structs.PartitionResp{"success"}
	bodyBytes, _ := json.Marshal(respBody)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

// Sends all View Info to new node
func sendUpdate(update structs.ViewUpdateForm) {
	log.Print("Sending Update to new Node")
	Ip := update.Ip
	Port := update.Port
	URL := "http://" + Ip + ":" + Port + "/viewchange"
	log.Print("url: "+URL)
	form := url.Values{}
	form.Add("my_Ip", Ip)
	for _, part := range _view {
		for _, node := range part {
			form.Add("Ip", node.Ip)
			form.Add("Port", node.Port)
			form.Add("Id", strconv.Itoa(node.Id))
			form.Add("Alive", strconv.FormatBool(node.Alive))
		}
		// if (v == len(_view)) {
		// 	form.Add("Ip", update.Ip)
		// 	form.Add("Port", update.Port)
		// 	form.Add("Id", strconv.Itoa(update.Id))
		// 	form.Add("Alive", strconv.FormatBool(update.Alive))
		// }
	}
	formJSON := form.Encode()
	req, _ := http.NewRequest(http.MethodPut, URL, strings.NewReader(formJSON))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	_, err := client.Do(req)
	ErrPanic(err)
	log.Print("I sent the goodz")
}

// Recreates View table and sets node Info
func AddNode(w http.ResponseWriter, r *http.Request) {
	log.Print("Filling out my Info cuz I'm a newbie")
	r.ParseForm()
	my_Ip := r.FormValue("my_Ip")
	Ip := r.PostForm["Ip"]
	Port := r.PostForm["Port"]
	Id := r.PostForm["Id"]
	Alive := r.PostForm["Alive"]
	/*K := r.FormValue("K")
	log.Print("Real K:" + K)
	_K,_ = strconv.Atoi(K)*/
	log.Print("My K:" + strconv.Itoa(_K))
	_view = make([][]structs.NodeInfo, int(math.Ceil(float64(len(Ip))/float64(_K))))
	for _, _ = range _view {
		_causal_Payload = append(_causal_Payload, 0)
	}
	log.Print("Length of View:" + string(len(_view)))
	for i, P := range Ip {
		log.Print("Ip: "+Ip[i])
		log.Print("Port: "+Port[i])
		log.Print("Id: "+Id[i])
		log.Print("Alive: "+Alive[i])
		id, _ := strconv.Atoi(Id[i])
		live, _ := strconv.ParseBool(Alive[i])
		temp := structs.NodeInfo{P, Port[i], id, live}
		if (my_Ip == P) {_my_node = temp}
		log.Print("My Node:")
		log.Print(_my_node)
		_view[id] = append(_view[id], temp)
	}
	log.Print("_my_node= ip:"+_my_node.Ip+" port:"+_my_node.Port+" id:"+strconv.Itoa(_my_node.Id)+" alive:"+strconv.FormatBool(_my_node.Alive))
	respBody := structs.PartitionResp{"success"}
	bodyBytes, _ := json.Marshal(respBody)
	w.WriteHeader(200)
	w.Write(bodyBytes)
	log.Print("View after adding myself")
	PrintView()
}

func SendKeyVal(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Print("In Send Key Val")
	Key := r.FormValue("key")
	Value := r.FormValue("val")
	log.Print("Key: " +Key + ", Val: " + Value)
	_causal_Payload[_my_node.Id]++
	//causal_payload := r.PostForm["causal_payload"]
	// do relevant kvs ops
	resp := _KVS.SetValue(Key, Value)
	// response preparation
	w.Header().Set("Content-Type", "application/json")
	if resp == "" {
		w.WriteHeader(201)
		// put.Replaced = 0
	} else {
		w.WriteHeader(200)
		// put.Replaced = 1
	}
	respBody := structs.PartitionResp{"success"}
	bodyBytes, _ := json.Marshal(respBody)
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

// new PUT handler for kvs manipulations
func NewSet(w http.ResponseWriter, r *http.Request) {
	log.Print("NewSet()")
	// log.Print("PUT request received")
	// var initialization
	put := structs.NewPUTResp{"success", _my_node.Id, StringifyPayload(), time.Now()}
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
		log.Print("Key Belongs to me")
		var resp string
		// check validitiy of key
		if !keyValid(putForm.Key) {
			// repsonse preparation
			put.Message = "error"
			// put.Replaced = 0
			// put.Owner = "undetermined"
			w.WriteHeader(401)
		} else {
			log.Print("Current View:")
			PrintView()
			_causal_Payload[_my_node.Id]++
			living := findAllLiving(_my_node)

			for _, node := range living {
				URL := "http://" + node.Ip + ":" + node.Port + "/sendKeyVal"
				form := url.Values{}
				form.Add("key", putForm.Key)
				form.Add("val", putForm.Value)
				/*for i := 0; i < len(putForm.Causal_Payload); i++ {
					form.Add("causal_payload", string(putForm.Causal_Payload[i]))
				}*/
				formJSON := form.Encode()
				// Request Creation
				req, _ := http.NewRequest(http.MethodPut, URL, strings.NewReader(formJSON))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				// Request sending logic
				client := &http.Client{
					Timeout: time.Second,
				}
				_, err := client.Do(req)
				ErrPanic(err)
			}

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
	log.Print("In NewGet")
	w.Header().Set("Content-Type", "application/json")
	// parse request for relevant data (key, get_number_of_keys)
	key, keyExists := r.URL.Query()["key"]
	// log.Print("key: "+key[0])
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
		log.Print("Key:" + key[0])
		resp := _KVS.GetValue(key[0])
		log.Print("Value:" + resp)
		if resp == "" {
			w.WriteHeader(404)
			getError := structs.ERROR{"error", "key does not exist"}
			jsonResponse, err = json.Marshal(getError)
		} else {
			get := structs.NewGETResp{"success", resp, _my_node.Id, StringifyPayload(), time.Now()}
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
	form := r.Form
	key := form["key"][0]
	log.Print("key: "+key)
	// URL logic
	URI := r.URL.RequestURI()
	index := partition.KeyBelongsTo(key, _view)
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
