package main

import (
	proto "Replica/gRPC"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

var isPrimary = flag.Bool("p", false, "")

var (
	echoNode      proto.NodeClient
	secondaryNode proto.NodeClient
	backupNodes   []proto.NodeClient
)

var ports []int
var (
	listeningPort int
	primaryPort   int
	secondaryPort int
)

var auctionOpenDuration = 180

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
	flag.Parse()
	SetLogger()
	RegisterPorts()
	RegisterNodes()

	node := StartNodeServer()
	StartListener(node)
}

func RegisterPorts() {
	// Parse each port
	for _, parameter := range os.Args[1:] {
		port, err := strconv.Atoi(parameter)
		if err != nil {
			//log.Printf("Invalid port '%s'. Please enter integers only.", parameter)
			continue
		}
		ports = append(ports, port)
	}

	// Check if there are enough ports
	if len(ports) < 1 {
		log.Print("Usage: go run Node.go <port1> <port2> ... <portN>")
		os.Exit(1)
	}

	// Print each port in the list
	fmt.Print("Ports: ")
	for _, port := range ports {
		fmt.Printf("%d ", port)
	}
	fmt.Println()

	listeningPort = ports[0]
	if !*isPrimary {
		primaryPort = ports[1]
	} else {
		secondaryIndex := rand.Intn(len(ports)-1) + 1
		secondaryPort = ports[secondaryIndex]
		log.Printf("Secondary: %d", secondaryPort)
	}
}

func RegisterNodes() {
	if !*isPrimary {
		echoNode = ConnectToNode(primaryPort)
	} else {
		secondaryNode = ConnectToNode(secondaryPort)
		for _, port := range ports {
			switch port {
			case listeningPort:
				continue
			case secondaryPort:
				continue
			}

			backupNode := ConnectToNode(port)
			backupNodes = append(backupNodes, backupNode)
		}
	}
}

func StartNodeServer() *NodeServer {
	nodeServer := &NodeServer{}

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

func ConnectToNode(port int) proto.NodeClient {
	portString := fmt.Sprintf(":%d", 16000+port)
	dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	connection, connectionEstablishErr := grpc.NewClient(portString, dialOptions)
	if connectionEstablishErr != nil {
		log.Fatalf("Could not establish connection on port %s | %v", portString, connectionEstablishErr)
	}

	log.Printf("Connected to node on port %s", portString)

	return proto.NewNodeClient(connection)
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

	var bidStatus int32
	if bid.Amount > nodeServer.highestBid {
		log.Printf("New bid %d, is winning bid! (previous highest: %d)", bid.Amount, nodeServer.highestBid)
		nodeServer.highestBid = bid.Amount
		log.Printf("New auction leader: nr. %d (Previous leader: nr. %d)", bid.Id, nodeServer.highestBidder)
		nodeServer.highestBidder = bid.Id
		bidStatus = bidSuccess
	} else {
		log.Printf("New bid %d, is less than winning bid %d, ignoring...", bid.Amount, nodeServer.highestBid)
		bidStatus = bidFail
	}

	err := nodeServer.Replicate()
	if err != nil {
		bidStatus = bidError
	}

	return &proto.BidAcknowledge{Status: bidStatus}, nil
}

func (nodeServer *NodeServer) Replicate() error {
	replicationData := &proto.AuctionData{
		HighestBid:     nodeServer.highestBid,
		HighestBidder:  nodeServer.highestBidder,
		AuctionEndTime: timestamppb.New(nodeServer.auctionEndTime), // Convert Go time to protobuf Timestamp
	}

	_, err := secondaryNode.ReplicateAuction(context.Background(), replicationData)
	if err != nil {
		return err
	}

	for _, backupNode := range backupNodes {
		go ReplicateToBackup(backupNode, replicationData)
	}

	return nil
}

func ReplicateToBackup(backupNode proto.NodeClient, replicationData *proto.AuctionData) {
	_, backupRepErr := backupNode.ReplicateAuction(context.Background(), replicationData)
	if backupRepErr != nil {
		// We don't care
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

func (nodeServer *NodeServer) ReplicateAuction(ctx context.Context, data *proto.AuctionData) (*proto.Empty, error) {
	nodeServer.highestBid = data.HighestBid
	nodeServer.highestBidder = data.HighestBidder
	nodeServer.auctionEndTime = data.AuctionEndTime.AsTime()

	log.Printf("Replication data received {%d by %d, ends at %v}", nodeServer.highestBid, nodeServer.highestBidder, nodeServer.auctionEndTime.Format("15:04:05"))

	return &proto.Empty{}, nil
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
