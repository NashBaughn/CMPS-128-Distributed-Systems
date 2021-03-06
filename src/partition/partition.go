package partition

import (
	"structs"
	"log"
	// "crypto/sha256"
	// "encoding/binary"
	"hash/fnv"
	"strconv"
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

func PrintRequestMap(reqMap map[int]*kvsAccess.KeyValStore) {
	log.Print("RequestMap:")
	for i, kvs := range(reqMap) {
		log.Print(strconv.Itoa(i)+"st partition: ")
		for k, v := range kvs.Store {
			log.Print("key: "+k+" value: "+v)
		}
	}
}

// main repartitioning Logic
// Generates map from IpPort to kvs.
func Repartition(index int, view [][]structs.NodeInfo, kvs *kvsAccess.KeyValStore) map[int]*kvsAccess.KeyValStore {
	log.Print("Repartition")
	// requestMap initialization
	// requestMap contains a mapping of keys to their correct partition #
	requestMap := make(map[int]*kvsAccess.KeyValStore)
    for i, part := range view {
      if (i != index && len(part) != 0) {
				requestMap[i] = kvsAccess.NewKVS()
			}
		/*for _, Head := range part {
			if Head.Alive != false {
				requestMap[Head.Ip+":"+Head.Port] = kvsAccess.NewKVS()
				break
			}
		}*/
	}
	for k, v := range kvs.Store {
		viewIndex := _hash(k, view)
		if viewIndex != index {
			kvs.DelValue(k)
			requestMap[viewIndex].SetValue(k, v)
		}
	}
	PrintRequestMap(requestMap)
	return requestMap
}
