package main

import (
    "mainInstance"
    //"networkMend"
    "github.com/gorilla/mux"
    "heartMonitor"
    "net/http"
    "log"
)

var hM = true

func main() {

  mainInstance.Start()
  if(hM) {go heartMonitor.Start()}
  log.Print("main_instance has started!")


  // create router instance
  router := mux.NewRouter()
  // designate handler funcs for each endpoint
  router.HandleFunc("/kvs", mainInstance.NewSet).Methods("POST", "PUT")
  router.HandleFunc("/kvs", mainInstance.NewGet).Methods("GET")
  router.HandleFunc("/kvs", mainInstance.NewDel).Methods("DELETE")
  router.HandleFunc("/kvs/view_update", mainInstance.SendViewUpdate).Methods("PUT")
  router.HandleFunc("/kvs/get_number_of_keys", mainInstance.NumKeys).Methods("GET")
  router.HandleFunc("/kvs/get_all_partition_ids", mainInstance.GetAllPartitionIds).Methods("GET")
  router.HandleFunc("/kvs/get_partition_id", mainInstance.GetPartitionId).Methods("GET")
  router.HandleFunc("/kvs/get_partition_members", mainInstance.GetPartitionMembers).Methods("GET")

  router.HandleFunc("/repartition", mainInstance.RepartitionHandler).Methods("PUT")
  router.HandleFunc("/partition", mainInstance.PartitionHandler).Methods("PUT")
  router.HandleFunc("/viewchange", mainInstance.AddNode).Methods("PUT")
  router.HandleFunc("/KVSMend", mainInstance.HandleKVSMend).Methods("PUT")
  router.HandleFunc("/heartbeat", mainInstance.HBresponse).Methods("GET")
  router.HandleFunc("/sendKeyVal", mainInstance.SendKeyVal).Methods("PUT")


  // listen on port 8080
  if err := http.ListenAndServe(":8080", router); err != nil {
      log.Fatal(err)
  }

}
