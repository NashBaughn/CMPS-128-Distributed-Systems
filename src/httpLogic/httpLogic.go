package httpLogic

import (
	"kvsAccess"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"structs"
)

var _ip = regexp.MustCompile(`(\d+\.){3}(\d+)`)
var _port = regexp.MustCompile(`\d{4}`)

// parses http.Request in view_update context
func ViewUpdateForm(r *http.Request) structs.ViewUpdateForm {
	r.ParseForm()
	IpPort := r.FormValue("ip_port")
	ip := _ip.FindString(IpPort)
	port := _port.FindString(IpPort)
	viewUpdateType := r.FormValue("type")
	viewUpdateForm := structs.ViewUpdateForm{viewUpdateType, ip, port}
	// Everything will be in the body
	//viewType, viewExists := r.URL.Query()["type"]
	//if viewExists { viewUpdateForm.Type = viewType[0] }
	return viewUpdateForm
}

// parses http.Request in PUT kvs context
func PutForm(r *http.Request) structs.PutForm {
	r.ParseForm()
	key := r.FormValue("key")
	value := r.FormValue("val")
	return structs.PutForm{key, value}
}

// parses http.Request in repartition context
func PartitionForm(r *http.Request) map[string]string {
    keyVal := make(map[string]string)
    //var kvArr []structs.KeyValue
	r.ParseForm()
	keys := r.PostForm["key"]
	vals := r.PostForm["val"]
	for i, key := range keys {
		keyVal[key] = vals[i]
	}
	return keyVal
}

type kv struct {
	key   string
	value string
}

// func CustomRequest(method string, header kv, URL string, formData []kv) *http.Request {
//   form := url.Values{}
//   for _, kv := range formData {
//     form.Add(kv.key, kv.value)
//   }
//   formJSON := form.Encode()
//   req, _ := http.NewRequest(http.MethodPut, URL, strings.NewReader(formJSON))
//   req.Header.Add(header.key, header.value)
//   return req
// }

// creates http.Request array to notify all nodes of view_update
func NotifyNodes(self structs.NodeInfo, viewForm structs.ViewUpdateForm, view [][]structs.NodeInfo) []*http.Request {
	var requestStore []*http.Request

	for _, part := range view {
        for _, node := range part {
    		tempIp := node.Ip
    		tempPort := node.Port
    		if self.Ip != tempIp {
    			URL := "http://" + tempIp + ":" + tempPort + "/partition"

    			// var formData = []kv {
    			//   kv { key  : "ip_port", value: viewForm.Ip+":"+viewForm.Port, },
    			//   kv { key  : "type", value: viewForm.Type, },
    			// }
    			// var header = kv { key  :"Content-Type", value:"application/x-www-form-urlencoded", }
    			// var req = CustomRequest("PUT", header, URL, formData)

    			form := url.Values{}
    			form.Add("ip_port", viewForm.Ip+":"+viewForm.Port)
    			form.Add("type", viewForm.Type)
    			formJSON := form.Encode()
    			req, _ := http.NewRequest(http.MethodPut, URL, strings.NewReader(formJSON))
    			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    			requestStore = append(requestStore, req)
            }
		}
	}
	return requestStore
}

// return http.Request array
// for sending all repartitioned keys to their corresponding nodes
func CreatePartitionRequests(view [][]structs.NodeInfo,
	requestMap map[int]*kvsAccess.KeyValStore) []*http.Request {
		
	var requestStore []*http.Request
	for ind, v := range requestMap {
		form := url.Values{}
		for k1, v1 := range v.Store {
			form.Add("key", k1)
			form.Add("val", v1)
		}
		formJSON := form.Encode()
		for _, node := range view[ind] {
			url := "http://" + node.Ip + ":" + node.Port + "/repartition"
			req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(formJSON))
			if err != nil {
				log.Panic(err)
			}
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			requestStore = append(requestStore, req)
		}
	}
	return requestStore
}
