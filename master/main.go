package main

import (
	masterGrpc "chaas/master/grpc"
	internal "chaas/master/internal"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"
)

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type change struct {
	hash   string
	nodeId string
}

type masterServer struct {
	mu sync.RWMutex
	masterGrpc.UnimplementedMasterServer
	hashToNodeId *map[string]string
	nodeIdToHash *map[string]string
	changeStream chan change
	clients      map[masterGrpc.Master_GetCacheStreamServer]struct{}
}

func NewmasterServer() *masterServer {
	hashToNodeId := map[string]string{}
	nodeIdToHash := map[string]string{}
	internal.LoadFromJson(&hashToNodeId, &nodeIdToHash)
	changeStream := make(chan change, 4)
	clients := map[masterGrpc.Master_GetCacheStreamServer]struct{}{}
	return &masterServer{hashToNodeId: &hashToNodeId, nodeIdToHash: &nodeIdToHash, changeStream: changeStream, clients: clients}
}

func (s *masterServer) GetCacheStream(req *masterGrpc.GetCacheStreamRequest, stream masterGrpc.Master_GetCacheStreamServer) error {
	s.mu.Lock()
	s.clients[stream] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, stream)
		s.mu.Unlock()
	}()
	s.mu.RLock()
	for hash, nodeId := range *s.hashToNodeId {
		stream.Send(&masterGrpc.GetCacheStreamResponse{Hash: hash, NodeId: nodeId})
	}
	s.mu.RUnlock()
	for change := range s.changeStream {
		s.mu.RLock()
		for client := range s.clients {
			go func(client masterGrpc.Master_GetCacheStreamServer) {
				err := client.Send(&masterGrpc.GetCacheStreamResponse{Hash: change.hash, NodeId: change.nodeId})
				if err != nil {
					log.Println("Error sending to client", err)
				}
			}(client)
		}
		s.mu.RUnlock()
	}
	return nil
}

func (s *masterServer) AddNode(ctx context.Context, req *masterGrpc.AddNodeRequest) (*masterGrpc.AddNodeResponse, error) {
	s.mu.Lock()
	nodeId := internal.AddNode(req.Hash, s.hashToNodeId, s.nodeIdToHash)
	s.mu.Unlock()
	s.changeStream <- change{hash: req.Hash, nodeId: nodeId}
	return &masterGrpc.AddNodeResponse{NodeId: nodeId}, nil
}

func (s *masterServer) RemoveNode(ctx context.Context, req *masterGrpc.RemoveNodeRequest) (*masterGrpc.RemoveNodeResponse, error) {
	s.mu.Lock()
	internal.RemoveNode(req.NodeId, s.hashToNodeId, s.nodeIdToHash)
	s.mu.Unlock()
	s.changeStream <- change{hash: (*s.nodeIdToHash)[req.NodeId], nodeId: "removed"}
	return &masterGrpc.RemoveNodeResponse{Success: true}, nil
}

func (s *masterServer) CreateHashRing(ctx context.Context, req *masterGrpc.CreateHashRingRequest) (*masterGrpc.CreateHashRingResponse, error) {

	s.mu.Lock()
	internal.CreateHashRing(req.NumNodes, s.hashToNodeId, s.nodeIdToHash)
	s.mu.Unlock()
	for hash, nodeId := range *s.hashToNodeId {
		s.changeStream <- change{hash: hash, nodeId: nodeId}
	}
	return &masterGrpc.CreateHashRingResponse{Success: true}, nil
}

func (s *masterServer) GetState(ctx context.Context, req *masterGrpc.GetStateRequest) (*masterGrpc.GetStateResponse, error) {
	return &masterGrpc.GetStateResponse{HashToNodeId: *s.hashToNodeId}, nil
}

func main() {
	fmt.Println("Welcome to Consistent Hashing as a Service - Master!")

	port := getEnv("MASTER_PORT", "50051")
	serverAddr := "0.0.0.0:" + port
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	masterGrpc.RegisterMasterServer(server, NewmasterServer())
	server.Serve(lis)
}

// func (s *masterServer) CreateHashRing(ctx context.Context, in *pb.CreateHashRingRequest) (*pb.CreateHashRingResponse, error) {
// 	log.Println("CreateHashRing", in)
// 	return &pb.CreateHashRingResponse{Success: true}, nil
// }
