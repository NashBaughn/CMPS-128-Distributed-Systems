package heartMonitor

import (
    "time"
    "net/http"
    "networkMend"
    "structs"
    "mainInstance"
)

func NewHeartMonitor() *HeartMonitor {
    return &HeartMonitor{}
}

type HeartMonitor struct {

}

func BeginMonitor(view [][]structs.NodeInfo) {
    for {
        time.Sleep(5000 * time.Millisecond)
        CheckNodes(view)
    }
}

func CheckNodes(view [][]structs.NodeInfo) {
    for _, row := range view {
        for _, node := range row {
            if(!SendPulse(node)){
                node.Alive = false
            } else {
                if(node.Alive == false){
                    networkMend.SendNetworkMend(node)
                    node.Alive = true
                }
            }

        }
    }
}

func SendPulse(node structs.NodeInfo) bool{
    URL := "http://" + node.Ip + ":" + node.Port + "/heartbeat"
    resp, err := http.Get(URL)
    mainInstance.ErrPanic(err)
    defer resp.Body.Close()
    // _, err = ioutil.ReadAll(resp.Body)
    if resp.StatusCode != 200 {
        return false
    }
    return true
}

// author: Alec
// purpose: placeholder for heartbeat request handling endpoint
// TODO: I added this so I could run our code against the test script
func HBresponse (w http.ResponseWriter, r *http.Request) {

}
