package main

import (
	proto "Replica/gRPC"
	"bufio"
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
	"strings"
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
	listeningPort = -1
	primaryPort   = -1
	secondaryPort = -1
	echoPort      = -1
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
	RegisterNodes()

	node := StartNodeServer()
	go StartListener(node)

	go ListenForInput()
}

// for testing crashes with "exit" command
func ListenForInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input := strings.ToLower(scanner.Text())
		commands := strings.Fields(input) // This splits the input by white space

		if len(commands) == 0 {
			log.Print("Usage: exit")
			continue
		} else if commands[0] == "exit" {
			os.Exit(0)
		}
	}
}

func RegisterNodes() {
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
	if len(ports) < 2 {
		log.Print("Usage (non primary): go run Node.go <currentPort> <primaryPort> <secondaryPort> ... <portN>")
		log.Print("Usage (Primary): go run Node.go <currentPort> <secondaryPort> ... <portN>")
		os.Exit(1)
	}

	// Print each port in the list
	fmt.Print("Ports: ")
	for _, port := range ports {
		fmt.Printf("%d ", port)
	}
	fmt.Println()

	listeningPort = ports[0] // port at index 0 is the port of the current node
	// if it is a normal node then the port at index 1 is the primary node port
	if !*isPrimary {
		primaryPort = ports[1]
		secondaryPort = ports[2]
		echoNode = ConnectToNode(primaryPort)
	} else { // if it is primary node then choose secondary and replicas
		ElectReplicas()
	}
}

func ElectReplicas() {
	ChooseNewSecondary()

	backupNodes = []proto.NodeClient{}
	for _, port := range ports {
		switch port { // skip current node port and secondary node port
		case listeningPort:
			continue
		case secondaryPort:
			continue
		}

		// add nodes as backup replicas if they are not current node or secondary node
		backupNode := ConnectToNode(port)
		backupNodes = append(backupNodes, backupNode)
	}
}

// choose a secondary node by taking the next port in port list
// called on app start or when old secondary node crashes
func ChooseNewSecondary() {
	secondaryIndex := rand.Intn(len(ports)-1) + 1
	if ports[secondaryIndex] == secondaryPort { // what
		ChooseNewSecondary()
		return
	}

	secondaryPort = ports[secondaryIndex]
	log.Printf("Secondary elected: %d", secondaryPort)
	secondaryNode = ConnectToNode(secondaryPort)
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

	echoPort = port
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
		log.Printf("Echo bid {ID: %d, Bid: %d} from %d to %d", bid.Id, bid.Amount, listeningPort, echoPort)
		bidAcknowledge, err := echoNode.Bid(ctx, bid)
		if err != nil {
			// this would occur if echoNode crashed, handle by connecting to secondary node
			log.Printf("Echo bid failed to node at port %d with error %d", echoPort, err.Error())
			echoNode = ConnectToNode(secondaryPort) // reconnect to secondary node instead
			return nodeServer.Bid(ctx, bid)         // Try bid again
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

	nodeServer.Replicate() // replicate data to other nodes

	return &proto.BidAcknowledge{Status: bidStatus}, nil
}

func (nodeServer *NodeServer) Replicate() {
	err := nodeServer.TryReplicate() // if this fails then elect a new secondary from backup nodes and try again
	if err != nil {                  // Secondary crashed
		ElectReplicas()        // Elect new secondary
		nodeServer.Replicate() // try again
	}
}

// replicate data with grpc to secondary node and backup nodes. if it fails with secondary node then return new error
func (nodeServer *NodeServer) TryReplicate() error {
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

// if this method calls that means this node is the secondary and it should track the primary node to know when it crashes
func (nodeServer *NodeServer) ReplicateAuction(ctx context.Context, data *proto.AuctionData) (*proto.Empty, error) {
	nodeServer.highestBid = data.HighestBid
	nodeServer.highestBidder = data.HighestBidder
	nodeServer.auctionEndTime = data.AuctionEndTime.AsTime()
	go monitorPort(primaryPort) // monitor primary node
	log.Printf("Replication data received {%d by %d, ends at %v}", nodeServer.highestBid, nodeServer.highestBidder, nodeServer.auctionEndTime.Format("15:04:05"))

	return &proto.Empty{}, nil
}

func monitorPort(port int) {
	portString := fmt.Sprintf("localhost:%d", 16000+port)
	for {
		conn, err := net.DialTimeout("tcp", portString, 1*time.Second)
		if err != nil {
			// primary node crashed, pretend to be primary somehow
		}
		err = conn.Close()
		if err != nil {
			return
		}
		time.Sleep(1 * time.Second) // Adjust interval as needed
	}
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
