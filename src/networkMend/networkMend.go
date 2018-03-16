package networkMend

import (
    "structs"
    "mainInstance"
    "kvsAccess"
    "strconv"
    "net/http"
    "net/url"
    "strings"
    "encoding/json"
)

// Sends Key Value Store to newly reconnected node in partition
// Also sends Causal Payload to check if KVS is up to date
// author: Alec
// update: first letter of function to upper case
// purpose: now it can be exported
func SendNetworkMend (Node structs.NodeInfo) {
    Ip := Node.Ip
	Port := Node.Port
	URL := "http://" + Ip + ":" + Port + "/networkMend"
	form := url.Values{}
	for key, val := range mainInstance.GetKVS().Store {
		form.Add("Key", key)
		form.Add("Val", val)
	}
    for _, val := range mainInstance.GetPayload() {
        form.Add("Payload", string(val))
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
func HandleNetworkMend (w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    Payload := r.PostForm["Payload"]
    newer := true
    var newPayload []int
    for ind, my_num := range mainInstance.GetPayload() {
        new_num,_  := strconv.Atoi(Payload[ind])
        if my_num > new_num {
            newer = false
            break
        }
        newPayload[ind] = new_num
    }
    if newer {
        newKVS := kvsAccess.NewKVS()
		keys := r.PostForm["Key"]
		vals := r.PostForm["Val"]
        for i, key := range keys {
            newKVS.SetValue(key, vals[i])
        }
        mainInstance.SetKVS(newKVS)
        mainInstance.SetPayload(newPayload)
    }
    respBody := structs.PartitionResp{"success"}
    bodyBytes, _ := json.Marshal(respBody)
    w.WriteHeader(200)
    w.Write(bodyBytes)
}
