package main

import (
    "mainInstance"
    "networkMend"
    "github.com/gorilla/mux"
)

func main() {

  mainInstance.Start()

  // create router instance
  router := mux.NewRouter()
  // designate handler funcs for each endpoint
  router.HandleFunc("/kvs", mainInstance.newSet).Methods("POST", "PUT")
  router.HandleFunc("/kvs", mainInstance.newGet).Methods("GET")
  router.HandleFunc("/kvs", mainInstance.newDel).Methods("DELETE")
  router.HandleFunc("/kvs/view_update", mainInstance.sendViewUpdate).Methods("PUT")
  router.HandleFunc("/repartition", mainInstance.repartitionHandler).Methods("PUT")
  router.HandleFunc("/partition", mainInstance.partitionHandler).Methods("PUT")
  router.HandleFunc("/viewchange", mainInstance.addNode).Methods("PUT")
  router.HandleFunc("/networkMend", networkMend.handleNetworkMend).Methods("PUT")
  router.HandleFunc("/kvs/get_number_of_keys", mainInstance.numKeys).Methods("GET")

  // listen on port 8080
  if err := http.ListenAndServe(":8080", router); err != nil {
      log.Fatal(err)
  }

}
