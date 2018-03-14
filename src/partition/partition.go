package partition

import (
	"structs"
	// "log"
	// "crypto/sha256"
	// "encoding/binary"
	"hash/fnv"
	// "strconv"
	// "math"
	"kvsAccess"
)

// checks if key belongs to specific ip:port
func KeyBelongs(key string, index int, view [][]structs.NodeInfo) bool {
	if _hash(key, view) == index {
		return true
	}
	return false
}

// checks which ip:port the key belongs to
func KeyBelongsTo(key string, view [][]structs.NodeInfo) int {
	return _hash(key, view)
}

// computes hash based on key
// returns that hash % len(view)
func _hash(key string, view [][]structs.NodeInfo) int {
	hash := fnv.New32a()
	hash.Write([]byte(key))
    var partitions []int
    for i, part := range view {
        if (len(part) != 0) {
            partitions = append(partitions, i)
        }
    }
	index := int(hash.Sum32()) % len(partitions)
	return partitions[index]
}

// main repartitioning Logic
// Generates map from IpPort to kvs.
func Repartition(index int, view [][]structs.NodeInfo, kvs *kvsAccess.KeyValStore) map[string]*kvsAccess.KeyValStore {
	// _ipPort := view[index].Ip+":"+view[index].Port
	// requestMap initialization
	// requestMap contains a mapping from all the keys to their new ip:port
	requestMap := make(map[string]*kvsAccess.KeyValStore)
    for i, part := range view {
        if (i != index && len(part) != 0) {
			requestMap[part[0].Ip+":"+part[0].Port] = kvsAccess.NewKVS()
		}
	}
	for k, v := range kvs.Store {
		viewIndex := _hash(k, view)
		if viewIndex != index {
			kvs.DelValue(k)
			node := view[viewIndex][0]
			requestMap[node.Ip+":"+node.Port].SetValue(k, v)
		}
	}
	return requestMap
}
