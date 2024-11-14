package main

import (
	proto "Replica/gRPC"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"strconv"
)

var ports []int

type NodeServer struct {
	proto.UnimplementedNodeServer
}

func main() {
	registerNodes()
	StartListener()
}

func registerNodes() {
	// Check if there are ports
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <port1> <port2> ... <portN>")
		os.Exit(1)
	}

	// Parse each port
	for _, parameter := range os.Args[1:] {
		port, err := strconv.Atoi(parameter)
		if err != nil {
			fmt.Printf("Invalid port '%s'. Please enter integers only.\n", parameter)
			continue
		}
		ports = append(ports, port)
	}

	// Print each port in the list?
	for _, port := range ports {
		fmt.Printf("Port: %d\n", port)
	}
}

func StartListener() {
	portString := fmt.Sprintf(":1600%d", ports[0])
	listener, listenErr := net.Listen("tcp", portString)
	if listenErr != nil {
		log.Fatalf("Failed to listen on port %s | %v", portString, listenErr)
	}

	grpcListener := grpc.NewServer()
	proto.RegisterNodeServer(grpcListener, &NodeServer{})

	log.Printf("Started listening on port %s", portString)

	serveListenerErr := grpcListener.Serve(listener)
	if serveListenerErr != nil {
		log.Fatalf("Failed to serve listener | %v", serveListenerErr)
	}
}

func (nodeServer *NodeServer) Bid(ctx context.Context, bid *proto.AuctionBid) (*proto.BidAcknowledge, error) {
	//TODO implement me
	panic("implement me")
}

func (nodeServer *NodeServer) Result(ctx context.Context, empty *proto.Empty) (*proto.AuctionOutcome, error) {
	//TODO implement me
	panic("implement me")
}
