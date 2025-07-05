package internal

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/google/uuid"
)

var hashFn = sha1.New()
var statePath = "master/state/"

func getDefaultHash(x int) *big.Int {
	hashBytes := []byte{}
	for range hashFn.Size() {
		hashBytes = append(hashBytes, byte(x))
	}
	hashDecimal := new(big.Int).SetBytes(hashBytes)
	return hashDecimal
}

func getMaxHash() *big.Int {
	return getDefaultHash(255)
}

func getMinHash() *big.Int {
	return getDefaultHash(0)
}

// TODO: need to optimize this
func saveToJson(hashToNodeId *map[string]string, nodeIdToHash *map[string]string) {
	os.MkdirAll(statePath, 0755)
	file, err := os.OpenFile(statePath+"hashToNodeId.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	json.NewEncoder(file).Encode(hashToNodeId)

	file, err = os.OpenFile(statePath+"nodeIdToHash.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	json.NewEncoder(file).Encode(nodeIdToHash)
}

func LoadFromJson(hashToNodeId *map[string]string, nodeIdToHash *map[string]string) {
	file, err := os.Open(statePath + "hashToNodeId.json")
	if err != nil {
		log.Println("no hashToNodeId.json found")
		return
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&hashToNodeId)

	file, err = os.Open(statePath + "nodeIdToHash.json")
	if err != nil {
		log.Println("no nodeIdToHash.json found")
		return
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&nodeIdToHash)
}

func getPartitionWidth(numNodes *big.Int) *big.Int {
	return new(big.Int).Div(getMaxHash(), numNodes)
}

func PrintJSON(obj interface{}) {
	bytes, _ := json.MarshalIndent(obj, "\t", "\t")
	fmt.Println(string(bytes))
}

func AddNode(hash string, hashToNodeId *map[string]string, nodeIdToHash *map[string]string) string {
	nodeId := uuid.New().String()
	(*hashToNodeId)[hash] = nodeId
	(*nodeIdToHash)[nodeId] = hash
	saveToJson(hashToNodeId, nodeIdToHash)
	return nodeId
}

func RemoveNode(nodeId string, hashToNodeId *map[string]string, nodeIdToHash *map[string]string) {
	delete(*hashToNodeId, (*nodeIdToHash)[nodeId])
	delete(*nodeIdToHash, nodeId)
	saveToJson(hashToNodeId, nodeIdToHash)
}

func CreateHashRing(numNodes string, hashToNodeId *map[string]string, nodeIdToHash *map[string]string) error {
	numNodesBigInt, ok := new(big.Int).SetString(numNodes, 10)
	// log.Println("numNodes", numNodes, numNodesBigInt, ok)
	if !ok {
		return fmt.Errorf("invalid numNodes")
	}
	if numNodesBigInt.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("numNodes must be greater than 0")
	}
	partitionWidth := getPartitionWidth(numNodesBigInt)
	start := getMinHash()
	bigInt1 := big.NewInt(1)
	clear(*hashToNodeId)
	clear(*nodeIdToHash)
	for i := big.NewInt(0); i.Cmp(numNodesBigInt) < 0; i.Add(i, bigInt1) {
		currentHash := start.Add(start, partitionWidth)
		currentNodeId := uuid.New().String()
		(*hashToNodeId)[currentHash.String()] = currentNodeId
		(*nodeIdToHash)[currentNodeId] = currentHash.String()
	}
	saveToJson(hashToNodeId, nodeIdToHash)
	return nil
}
