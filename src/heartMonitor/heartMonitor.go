package heartMonitor

import (
    "time"
    "net/http"
    "log"
    //"networkMend"
    "structs"
)


func Start(curr structs.NodeInfo, view *[][]structs.NodeInfo) {
    for {
        time.Sleep(500 * time.Millisecond) // 500 ms for production
        log.Print("HeartBeat")
        log.Print(curr.Ip)
        log.Print(*view)
        CheckNodes(*view, curr.Ip)
    }
}

func CheckNodes(view [][]structs.NodeInfo, Ip string) {
    for i, row := range view {
        for j, node := range row {
            if (Ip != node.Ip){
                if(!SendPulse(node)){
                    log.Print("Dead node: "+node.Ip)
                    view[i][j].Alive = false
                } else {
                    if(node.Alive == false){
                        log.Print("Resurrected node: "+node.Ip)
                        //networkMend.SendNetworkMend(node)
                        view[i][j].Alive = true
                    }
                }
            }

        }
    }
}

func SendPulse(node structs.NodeInfo) bool{
    URL := "http://" + node.Ip + ":" + node.Port + "/heartbeat"
    //log.Print(URL)
    resp, err := http.Get(URL)
    timeout := time.Duration(1 * time.Second)
    client := http.Client{
        Timeout: timeout,
    }
    client.Get(URL)
    if err != nil{
        log.Print(err)
        return false
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return false
    }
    return true
}
