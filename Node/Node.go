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
var secondaryPort int

var auctionOpenDuration = 100

var ports []int
var listeningPort, echoPort int
var echoNode proto.NodeClient

var (
	bidSuccess int32 = 0
	bidFail    int32 = 1
	bidError   int32 = 2
)

type NodeServer struct {
	proto.UnimplementedNodeServer
	highestBid     int32
	highestBidder  int32
	auctionEndTime time.Time
}

func main() {
	SetLogger()
	registerNodes()
	node := StartNodeServer()

	StartSender()
	StartListener(node)
}

func registerNodes() {
	// Parse each port
	for _, parameter := range os.Args[1:] {
		port, err := strconv.Atoi(parameter)
		if err != nil {
			//log.Printf("Invalid port '%s'. Please enter integers only.", parameter)
			continue
		}
		ports = append(ports, port)
	}

	// Check if there are ports
	if len(ports) < 1 {
		log.Print("Usage: go run Node.go <port1> <port2> ... <portN>")
		os.Exit(1)
	}

	// Print each port in the list?
	fmt.Print("Ports: ")
	for _, port := range ports {
		fmt.Printf("%d ", port)
	}
	fmt.Println()

	listeningPort = ports[0]
	echoPort = ports[1]
}

func StartNodeServer() *NodeServer {
	nodeServer := &NodeServer{}

	flag.Parse()
	if *isPrimary {
		go nodeServer.CloseAuctionAfter(auctionOpenDuration)
	}

	return nodeServer
}

func (nodeServer *NodeServer) CloseAuctionAfter(nSeconds int) {
	log.Printf("Auction will close in %v seconds", nSeconds)
	auctionDuration := time.Duration(nSeconds) * time.Second
	nodeServer.auctionEndTime = time.Now().Add(auctionDuration)
	time.Sleep(time.Duration(nSeconds) * time.Second)
	log.Printf("Auction has closed, highest bid was %d by bidder nr. %d", nodeServer.highestBid, nodeServer.highestBidder)
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
		if nodeServer.IsAuctionClosed() {
			log.Printf("Bidder nr. %d, just tried to bid on a closed auction, what a nerd", bid.Id)
			return &proto.BidAcknowledge{Status: bidFail}, nil
		}
		return nodeServer.MakeBid(bid)
	} else {
		log.Printf("Echo bid {ID: %d, Bid: %d}", bid.Id, bid.Amount)
		bidAcknowledge, err := echoNode.Bid(ctx, bid)
		if err != nil {
			return &proto.BidAcknowledge{Status: bidError}, err
		}
		log.Printf("Response {Status: %d}", bidAcknowledge.Status)
		return bidAcknowledge, nil
	}
}

func (nodeServer *NodeServer) MakeBid(bid *proto.AuctionBid) (*proto.BidAcknowledge, error) {
	log.Printf("Bidder nr. %d, just bid %d", bid.Id, bid.Amount)
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
		return &proto.AuctionOutcome{IsFinished: nodeServer.IsAuctionClosed(), HighestBid: nodeServer.highestBid, LeaderId: nodeServer.highestBidder}, nil
	} else {
		log.Printf("Echo result")
		auctionOutcome, err := echoNode.Result(ctx, empty)
		if err != nil {
			return nil, err
		}

		log.Printf("Response {Closed: %v, Bid: %d by %d}", auctionOutcome.IsFinished, auctionOutcome.HighestBid, auctionOutcome.LeaderId)
		return auctionOutcome, nil
	}
}

func (nodeServer *NodeServer) IsAuctionClosed() bool {
	return time.Now().After(nodeServer.auctionEndTime)
}

type logWriter struct {
}

func SetLogger() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("15:04:05.9 ") + string(bytes))
}
