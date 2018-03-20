package heartMonitor

import (
    "time"
    "net/http"
    "log"
    "mainInstance"
    "structs"
    "strings"
)


func Start() {
    log.Print("Starting Heart Monitor")
    for {
        time.Sleep(100 * time.Millisecond) // 500 ms for production
        //log.Print("HeartBeat")
        view := mainInstance.GetView()
        currNode := mainInstance.GetNode()
        //log.Print(curr.Ip)
        //mainInstance.PrintView()
        CheckNodes(view, currNode)
    }
}

func CheckNodes(view [][]structs.NodeInfo, currNode structs.NodeInfo) {
    for i, row := range view {
        for j, node := range row {
            if (currNode.Ip != node.Ip){
                if(!SendPulse(node)) {
                    log.Print("Dead node: "+node.Ip)
                    view[i][j].Alive = false
                } else {
                    if(node.Alive == false){
                        log.Print("Resurrected node: "+node.Ip)
                        if (currNode.Id == node.Id) {
                            mainInstance.SendKVSMend(node)
                        }
                        view[i][j].Alive = true
                    }
                }
            }
        }
    }
}

func SendPulse(node structs.NodeInfo) bool{
    URL := "http://" + node.Ip + ":" + node.Port + "/heartbeat"
    // Request Creation
    req, _ := http.NewRequest(http.MethodGet, URL, strings.NewReader(""))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    // Request sending logic
    client := &http.Client{
        Timeout: 500 * time.Millisecond,
    }
    _, err := client.Do(req)
    if err != nil{
        log.Print(err)
        return false
    }
    /*if resp.StatusCode != 200 {
        return false
    }*/
    return true
}
