package structs
import (
	"time"
)

// File full of data structures

type NodeInfo struct {
	Ip   string
	Port string
	Id   int
	Alive bool
}

// author: Reese & Alec
// purpose: struct that holds a node's personal, internal state
// add Causal_payload to account for causal ordering
type NewNodeInfo struct {
	Ip   string
	Port string
	Id   int
	Causal_Payload []int
}

type ViewUpdateForm struct {
	Type string
	Ip   string
	Port string
}

type PutForm struct {
	Key   string
	Value string
}

// author: Alec
// scenario: a node receives a PUT request and needs to parse the form body
// purpose: provides template for parsed form body of a PUT request
// added the Causal_Payload field to account for causal ordering
type NewPutForm struct {
	Key   string
	Value string
	Causal_Payload []int
}

type PartitionRequest struct {
	Ip   string
	Port string
}

type KeyValue struct {
	Key   string
	Value string
}

type PartitionResp struct {
	Message string `json:"msg"`
}

// author: Alec
// scenario: node receives a request from a client to add a node
// purpose: response body template for the node addition request
type ViewUpdateAddResp struct {
	Message string `json:"msg"`
	Partition_Id int `json:"partition_id"`
	Number_Of_Partitions int `json:"number_of_partitions"`
}

// author: Alec
// scenario: node receives a request from a client to remove a node
// purpose: response body template for the node removal request
type ViewUpdateRemoveResp struct {
	Message string `json:"msg"`
	Number_Of_Partitions int `json:"number_of_partitions"`
}

type NumKeys struct {
	Count int `json:"count"`
}

// author: Alec
// scenario: node receives a request asking for the partition_id of itself
// purpose: response body template for the partition_id request
type GETPartitionIdResp struct {
	Message string `json:"msg"`
	Partition_Id int `json:"partition_id"`
}

// author: Alec
// scenario: node receives a request asking for a list of partition_ids of the partition it is in
// purpose: response body template for the partition_id request
type GETAllPartitionIdsResp struct {
	Message string `json:"msg"`
	Partition_Id_List []int `json:"partition_id_list"`
}

// author: Alec
// scenario: node receives a request asking for the IpPorts of the partition it belongs to
// purpose: response body template for the partition_id request
type GETPartitionMembersResp struct {
	Message string `json:"msg"`
	Partition_Members []int `json:"partition_members"`
}

type GET struct {
	Message string `json:"msg"`
	Value   string `json:"value"`
	Owner   string `json:"owner"`
}

// author: Alec
// scenario: node receives a GET request and needs to send a response back
// purpose: response body template for GET request
// TODO: Do we want Timestamp to be an int?
type newGETResp struct {
	Message string `json:"msg"`
	Value   string `json:"value"`
	Partition_Id int `json:"partition_id"`
	Causal_Payload []int `json:"causal_payload"`
	Timestamp time.Time `json:"timestamp"` // do we want this as an int?
}

type PUT struct {
	Replaced int    `json:"replaced"`
	Message  string `json:"msg"`
	Owner    string `json:"owner"`
}

// author: Alec
// scenario: node receives a PUT request and needs to send a response back
// purpose: response body template for PUT request
// TODO: Do we want Timestamp to be an int?
type NewPUTResp struct {
	Message string `json:"msg"`
	Partition_Id int `json:"partition_id"`
	Causal_Payload []int `json:"causal_payload"`
	Timestamp time.Time `json:"timestamp"` // do we want this as an int?
}

type ERROR struct {
	Message string `json:"msg"`
	Error   string `json:"error"`
}

type DELExists struct {
	Message string `json:"msg"`
	Owner   string `json:"owner"`
}
