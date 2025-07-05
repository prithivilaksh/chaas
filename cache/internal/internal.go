package internal

import (
	"crypto/sha1"
	"math/big"
	"sort"
)

var hashFn = sha1.New()

func getSha1Hash(key string) *big.Int {
	hashFn.Reset()
	hashFn.Write([]byte(key))
	hashBytes := hashFn.Sum(nil)
	hashDecimal := new(big.Int).SetBytes(hashBytes)
	return hashDecimal
}

func getNodeIdByHash(hash string, hashToNodeId *map[string]string) string {
	return (*hashToNodeId)[hash]
}

func getNextNearestNodehash(keyHash *big.Int, hashToNodeId *map[string]string) string {
	hashSlice := []string{}
	for key := range *hashToNodeId {
		hashSlice = append(hashSlice, key)
	}

	sort.Strings(hashSlice)
	l, r, pos := 0, len(hashSlice)-1, 0
	for l <= r {
		mid := l + (r-l)/2
		if keyHash.String() <= hashSlice[mid] {
			pos = mid
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return hashSlice[pos]
}

func GetNextNodeIdByKey(key string, hashToNodeId *map[string]string) string {
	keyHash := getSha1Hash(key)
	nodeHash := getNextNearestNodehash(keyHash, hashToNodeId)
	return getNodeIdByHash(nodeHash, hashToNodeId)
}

func UpdateCache(nodeId string, hash string, hashToNodeId *map[string]string) {
	(*hashToNodeId)[hash] = nodeId
}
