package main

import (
	cacheGrpc "chaas/cache/grpc"
	internal "chaas/cache/internal"
	masterGrpc "chaas/master/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type cacheServer struct {
	cacheGrpc.UnimplementedCacheServer
	hashToNodeId *map[string]string
}

func initCache(hashToNodeId *map[string]string) {
	masterHost := getEnv("MASTER_HOST", "master")
	masterPort := getEnv("MASTER_PORT", "50051")
	masterAddr := fmt.Sprintf("%s:%s", masterHost, masterPort)

	conn, err := grpc.Dial(masterAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to master: %v", err)
	}
	defer conn.Close()

	masterClient := masterGrpc.NewMasterClient(conn)
	stream, err := masterClient.GetCacheStream(context.Background(), &masterGrpc.GetCacheStreamRequest{})
	if err != nil {
		log.Fatalf("failed to get cache stream: %v", err)
	}
	defer stream.CloseSend()

	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Fatalf("failed to receive from cache stream: %v", err)
		}
		(*hashToNodeId)[resp.Hash] = resp.NodeId
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func NewCacheServer() *cacheServer {
	hashToNodeId := map[string]string{}
	go initCache(&hashToNodeId)

	return &cacheServer{hashToNodeId: &hashToNodeId}
}

func (s *cacheServer) GetNextNodeIdByKey(ctx context.Context, req *cacheGrpc.GetNextNodeIdByKeyRequest) (*cacheGrpc.GetNextNodeIdByKeyResponse, error) {
	nodeId := internal.GetNextNodeIdByKey(req.Key, s.hashToNodeId)
	if nodeId == "removed" {
		return nil, status.Errorf(codes.NotFound, "nodeId not found")
	}
	return &cacheGrpc.GetNextNodeIdByKeyResponse{NodeId: nodeId}, nil
}

func (s *cacheServer) UpdateCache(ctx context.Context, req *cacheGrpc.UpdateCacheRequest) (*cacheGrpc.UpdateCacheResponse, error) {
	internal.UpdateCache(req.NodeId, req.Hash, s.hashToNodeId)
	return &cacheGrpc.UpdateCacheResponse{Success: true}, nil
}

func (s *cacheServer) GetState(ctx context.Context, req *cacheGrpc.GetStateRequest) (*cacheGrpc.GetStateResponse, error) {
	return &cacheGrpc.GetStateResponse{HashToNodeId: *s.hashToNodeId}, nil
}

func main() {
	fmt.Println("Welcome to Consistent Hashing as a Service - Cache!")

	port := getEnv("CACHE_PORT", "50052")
	serverAddr := "0.0.0.0:" + port
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	cacheGrpc.RegisterCacheServer(grpcServer, NewCacheServer())
	grpcServer.Serve(lis)
}
