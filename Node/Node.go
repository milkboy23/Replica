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
	"time"
)

var ports []int
var auctionOpen bool
var openTime = 100

var (
	bidSuccess int32 = 0
	bidFail    int32 = 1
	bidError   int32 = 2
)

type NodeServer struct {
	proto.UnimplementedNodeServer
	bidders       map[int32]int32
	highestBid    int32
	highestBidder int32
}

func main() {
	registerNodes()
	node := NewNodeServer()
	go node.CloseAuctionAfter(openTime)
	node.StartListener()
}

func (nodeServer *NodeServer) CloseAuctionAfter(nSeconds int) {
	auctionOpen = true
	log.Printf("Auction will close in %v seconds", nSeconds)
	time.Sleep(time.Duration(nSeconds) * time.Second)
	auctionOpen = false
	log.Printf("Auction has closed, highest bid was %d by bidder nr. %d", nodeServer.highestBid, nodeServer.highestBidder)
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

func NewNodeServer() *NodeServer {
	return &NodeServer{
		bidders: make(map[int32]int32),
	}
}

func (nodeServer *NodeServer) StartListener() {
	portString := fmt.Sprintf(":1600%d", ports[0])
	listener, listenErr := net.Listen("tcp", portString)
	if listenErr != nil {
		log.Fatalf("Failed to listen on port %s | %v", portString, listenErr)
	}

	grpcListener := grpc.NewServer()
	proto.RegisterNodeServer(grpcListener, nodeServer)

	log.Printf("Started listening on port %s", portString)

	serveListenerErr := grpcListener.Serve(listener)
	if serveListenerErr != nil {
		log.Fatalf("Failed to serve listener | %v", serveListenerErr)
	}
}

func (nodeServer *NodeServer) Bid(ctx context.Context, bid *proto.AuctionBid) (*proto.BidAcknowledge, error) {
	if !auctionOpen {
		log.Printf("Bidder nr. %d, just tried to bid on a closed auction, what a nerd", bid.Id)
		return &proto.BidAcknowledge{Status: bidFail}, nil
	}

	log.Printf("Bidder nr. %d, just bid %d", bid.Id, bid.Amount)
	bidderHighestBid, bidderAlreadyRegistered := nodeServer.bidders[bid.Id]
	if !bidderAlreadyRegistered {
		nodeServer.bidders[bid.Id] = bid.Amount
		log.Print("Bidder wasn't registered, registering bidder...")
	} else {
		log.Printf("Bidder already registered with highest bid %d", bidderHighestBid)
		if bid.Amount > bidderHighestBid {
			nodeServer.bidders[bid.Id] = bid.Amount
			log.Print("Bidder's new bid, is higher than previous highest bid, updating...")
		}
	}

	if bid.Amount > nodeServer.highestBid {
		log.Printf("New bid %d, is winning bid! (previous highest: %d)", bid.Amount, nodeServer.highestBid)
		nodeServer.highestBid = bid.Amount
		log.Printf("New auction leader: nr. %d (Previous leader: nr. %d)", bid.Id, nodeServer.highestBidder)
		nodeServer.highestBidder = bid.Id
		return &proto.BidAcknowledge{Status: bidSuccess}, nil
	} else {
		log.Printf("New bid %d, is less than winning bid %d, ignoring...", bid.Amount, nodeServer.highestBid)
		return &proto.BidAcknowledge{Status: bidFail}, nil
	}
}

func (nodeServer *NodeServer) Result(ctx context.Context, empty *proto.Empty) (*proto.AuctionOutcome, error) {
	return &proto.AuctionOutcome{IsAuctionFinished: !auctionOpen, HighestBid: nodeServer.highestBid, WinningBidder: nodeServer.highestBidder}, nil
}
