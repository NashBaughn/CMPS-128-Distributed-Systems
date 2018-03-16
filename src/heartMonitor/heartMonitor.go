package heartMonitor

import (
    "time"
    "net/http"
    "structs"
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

func CheckNodes(node NodeInfo, view [][]structs.NodeInfo) {
    for _, row := range view {
        for _, node := range row {
            if(!SendPulse(node)){
                node.alive = false
            } else {
                if(node.alive == false){
                    partitionMend(node)
                    node.alive = true
                }
            }

        }
    }
}

func SendPulse(node NodeInfo) bool{
    URL := "http://" + node.Ip + ":" + node.Port + "/heartbeat"
    resp, err := http.Get(URL)
    if err != nil {
        // handle error
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if resp.StatusCode != 200 {
        return false
    }
    return true
}



