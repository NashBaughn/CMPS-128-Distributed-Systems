package structs

// File full of data structures

// Changed this from IpPort to include more info
type NodeInfo struct {
	Ip   string
	Port string
	Id   int
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

type NumKeys struct {
	Count int `json:"count"`
}

type GET struct {
	Message string `json:"msg"`
	Value   string `json:"value"`
	Owner   string `json:"owner"`
}

type PUT struct {
	Replaced int    `json:"replaced"`
	Message  string `json:"msg"`
	Owner    string `json:"owner"`
}

type ERROR struct {
	Message string `json:"msg"`
	Error   string `json:"error"`
}

type DELExists struct {
	Message string `json:"msg"`
	Owner   string `json:"owner"`
}
