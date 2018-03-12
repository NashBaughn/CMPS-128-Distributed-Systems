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
func KeyBelongs(key string, index int, view []structs.IpPort) bool {
	if _hash(key, view) == index {
		return true
	}
	return false
}

// checks which ip:port the key belongs to
func KeyBelongsTo(key string, view []structs.IpPort) int {
	return _hash(key, view)
}

// computes hash based on key
// returns that hash % len(view)
func _hash(key string, view []structs.IpPort) int {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	index := int(hash.Sum32()) % len(view)
	return index
}

// main repartitioning Logic
// handles node addiition and deletion
// Needs to be changed!!
func Repartition(index int, view []structs.IpPort, kvs *kvsAccess.KeyValStore) map[string]*kvsAccess.KeyValStore {
	// _ipPort := view[index].Ip+":"+view[index].Port
	// requestMap initialization
	// requestMap contains a mapping from all the keys to their new ip:port
	requestMap := make(map[string]*kvsAccess.KeyValStore)
	for i := 0; i < len(view); i++ {
		if index != i {
			requestMap[view[i].Ip+":"+view[i].Port] = kvsAccess.NewKVS()
		}
	}
	for k, v := range kvs.Store {
		viewIndex := _hash(k, view)
		if viewIndex != index {
			kvs.DelValue(k)
			ipPort := view[viewIndex]
			requestMap[ipPort.Ip+":"+ipPort.Port].SetValue(k, v)
		}
	}
	return requestMap
}
