package main

import (
	proto "Replica/gRPC"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var isPrimary = flag.Bool("p", false, "")

var ports []int
var listeningPort, echoPort int
var echoNode proto.NodeClient

var auctionOpen = true
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
	flag.Parse()
	registerNodes()
	node := NewNodeServer()

	if *isPrimary {
		go node.CloseAuctionAfter(openTime)
	}

	StartSender()
	StartListener(node)
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
		log.Print("Usage: go run client.go <port1> <port2> ... <portN>")
		os.Exit(1)
	}

	// Parse each port
	for _, parameter := range os.Args[1:] {
		port, err := strconv.Atoi(parameter)
		if err != nil {
			//log.Printf("Invalid port '%s'. Please enter integers only.", parameter)
			continue
		}
		ports = append(ports, port)
	}

	// Print each port in the list?
	for _, port := range ports {
		log.Printf("Port: %d", port)
	}

	listeningPort = ports[0]
	echoPort = ports[1]
}

func NewNodeServer() *NodeServer {
	return &NodeServer{
		bidders: make(map[int32]int32),
	}
}

func StartSender() {
	portString := fmt.Sprintf(":%d", 16000+echoPort)
	dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	connection, connectionEstablishErr := grpc.NewClient(portString, dialOptions)
	if connectionEstablishErr != nil {
		log.Fatalf("Could not establish connection on port %s | %v", portString, connectionEstablishErr)
	}

	log.Printf("Started sending on port %s", portString)

	echoNode = proto.NewNodeClient(connection)
}

func StartListener(nodeServer *NodeServer) {
	portString := fmt.Sprintf(":%d", 16000+listeningPort)
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
	if *isPrimary {
		return nodeServer.MakeBid(bid)
	} else {
		bidAcknowledge, err := echoNode.Bid(ctx, bid)
		if err != nil {
			return &proto.BidAcknowledge{Status: bidError}, err
		}
		return bidAcknowledge, nil
	}
}

func (nodeServer *NodeServer) MakeBid(bid *proto.AuctionBid) (*proto.BidAcknowledge, error) {
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
	if *isPrimary {
		return &proto.AuctionOutcome{IsAuctionFinished: !auctionOpen, HighestBid: nodeServer.highestBid, WinningBidder: nodeServer.highestBidder}, nil
	} else {
		auctionOutcome, err := echoNode.Result(ctx, empty)
		if err != nil {
			return nil, err
		}

		return auctionOutcome, nil
	}
}
