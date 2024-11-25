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

var (
	isPrimary     bool                        // Is this node the primary?
	isPrimaryFlag = flag.Bool("p", false, "") // Simple flag to set on startup if a node is primary
)

var (
	echoNode      proto.NodeClient   // The node to echo commands to, if this node is an echo node
	secondaryNode proto.NodeClient   // A reference to the secondary, used only by the primary node
	backupNodes   []proto.NodeClient // A list of backup nodes, used only by the primary node
)

var ports []int // An array/slice of ports given on startup
var (
	listeningPort = -1 // The port this node listens on
	primaryPort   = -1 // The port of the primary node, used only by echo nodes
	secondaryPort = -1 // The port of the secondary node, used by all nodes
	echoPort      = -1 // The port to echo commands to if this is an echo node
)

var auctionOpenDuration = 120 // The duration of the auction in seconds

var (
	bidSuccess int32 = 0 // Constant signifying a bid was successful
	bidFail    int32 = 1 // Constant signifying a bid failed
	bidError   int32 = 2 // Constant signifying there was an error when trying to bid
)

// NodeServer A reference to the node on which the auction runs
type NodeServer struct {
	proto.UnimplementedNodeServer
	highestBid     int32     // The current highest bid
	highestBidder  int32     // The ID of the current highest bidder
	auctionEndTime time.Time // The time at which the auction ends
}

func main() {
	flag.Parse()
	isPrimary = *isPrimaryFlag // Sets whether this node is primary from flag

	SetLogger()     // Method for setting custom log.print() format
	RegisterNodes() // Connects to nodes

	node := StartNodeServer()
	StartListener(node)
}

// RegisterNodes Sets and connects to all relevant ports/nodes
func RegisterNodes() {
	// Parse each launch argument from string to (port) integer
	for _, parameter := range os.Args[1:] {
		port, err := strconv.Atoi(parameter)
		if err != nil {
			continue // We just ignore anything that isn't a valid port
		}
		ports = append(ports, port)
	}

	// Check if there are enough ports supplied, always at least 3 required (own, primary, secondary, backups. Some configuration of these)
	if len(ports) < 3 {
		log.Print("Usage (Backup): go run Node.go <currentPort> <primaryPort> <secondaryPort>")
		log.Print("Usage (Secondary): go run Node.go <currentPort> <primaryPort> <secondaryPort/currentPort> <port1> ... <portN>")
		log.Print("Usage (Primary): go run Node.go <currentPort> <secondaryPort> <port1> ... <portN>")
		os.Exit(1)
	}

	// Print each port in the list
	fmt.Print("Ports: ")
	for _, port := range ports {
		fmt.Printf("%d ", port)
	}
	fmt.Println()

	listeningPort = ports[0] // Port at index 0 is the port of the current node
	if !isPrimary {          // If this node isn't the primary node then
		primaryPort = ports[1]         // The port at index 1, is the primary node port
		secondaryPort = ports[2]       // The port at index 2, is the secondary node port
		ConnectToEchoNode(primaryPort) // Sets our echo node to the primary
	} else { // If this node is the primary node then
		secondaryPort = ports[1] // The port at index 1, is the secondary node port
		ConnectToSecondary()     // Connects to our secondary node
		ElectReplicas()          // Sets any remaining ports, as backup nodes
	}
}

// ConnectToEchoNode Sets the echo node and port to use the given port
func ConnectToEchoNode(port int) {
	echoPort = port                // Updates echo port to given port
	echoNode = ConnectToNode(port) // Updates echoNode to node with given port
}

// ConnectToSecondary Updates the secondaryNode to node with secondaryPort
func ConnectToSecondary() {
	secondaryNode = ConnectToNode(secondaryPort)
}

// ElectReplicas Sets all unused ports as backup nodes
func ElectReplicas() {
	backupNodes = []proto.NodeClient{} // Clears backupNodes array
	for _, port := range ports {
		if port == listeningPort || port == secondaryPort || port == primaryPort { // Skip our own node port, and the primary and secondary node ports
			continue
		}

		// Add remaining nodes as backup replicas
		backupNode := ConnectToNode(port)
		backupNodes = append(backupNodes, backupNode)
	}
}

// ElectNewSecondary Elects a new secondary and updates replicas accordingly
func ElectNewSecondary() {
	randomPort := rand.Intn(len(ports)-1) + 1 // Goes from 1 to len(ports), so that we skip our listening port

	// Checks if our random port is accidentally the same as the old one,
	// or in case of the secondary promoting, that it isn't the old primary (which has crashed in this case)
	if ports[randomPort] == secondaryPort || ports[randomPort] == primaryPort {
		ElectNewSecondary() // If so, try again
		return
	}

	secondaryPort = ports[randomPort] // Finally, a secondary port that works :D
	log.Printf("Secondary elected: %d", secondaryPort)
	ConnectToSecondary() // Set out secondary node to use this new port
	ElectReplicas()      // Re-elect replicas
}

// StartNodeServer Starts a new node server and boots up auction if this node is primary
func StartNodeServer() *NodeServer {
	if isPrimary { // Sets up auction state on primary (initial bid being 0, initial bidder being no one (ID: -1), and the auction end time)
		auctionDuration := time.Duration(auctionOpenDuration) * time.Second // Converts from int seconds to time.Duration
		nodeServer := &NodeServer{
			highestBid:     0,                               // Initial bid of 0
			highestBidder:  -1,                              // Symbolizes that we initially have no highest bidder (ID: -1)
			auctionEndTime: time.Now().Add(auctionDuration), // Sets auctionEndTime to be auctionDuration seconds from now
		}

		nodeServer.Replicate()            // Ensures all other nodes initially have this same data ^^
		go nodeServer.CloseAuctionAfter() // Starts auction (goroutine with announcing open/close)

		return nodeServer
	} else { // For non-primary nodes
		return &NodeServer{} // They have their data overwritten from initial replication called by code above anyway, so empty is fine
	}
}

// CloseAuctionAfter Calculates remaining time of auction, and prints said remaining time. Also sleeps until auction ends and prints winner
// WARNING: THIS CALLS SLEEP, SO USE ONLY IN GOROUTINE
func (nodeServer *NodeServer) CloseAuctionAfter() {
	timeRemaining := nodeServer.auctionEndTime.Sub(time.Now())                   // auctionEndTime - time.Now = remainingTime
	log.Printf("Auction will close in %v seconds", int(timeRemaining.Seconds())) // Round to integer
	time.Sleep(timeRemaining)                                                    // Waits until auction closes
	log.Printf("Auction has closed, highest bid was %d by bidder nr. %d", nodeServer.highestBid, nodeServer.highestBidder)
}

// ConnectToNode Connects to, and returns, the node at the given port
func ConnectToNode(port int) proto.NodeClient {
	portString := fmt.Sprintf(":%d", 16000+port) // Formats port string
	dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	connection, connectionEstablishErr := grpc.NewClient(portString, dialOptions)
	if connectionEstablishErr != nil {
		log.Fatalf("Could not establish connection on port %s | %v", portString, connectionEstablishErr)
	}

	log.Printf("Connected to node on port %s", portString) // This can get obnoxious, just remove if it's annoying

	return proto.NewNodeClient(connection)
}

// StartListener Starts the listener that receives gRPC calls
func StartListener(nodeServer *NodeServer) {
	portString := fmt.Sprintf(":%d", 16000+listeningPort) // Formats port string
	listener, listenErr := net.Listen("tcp", portString)  // Starts listener
	if listenErr != nil {
		log.Fatalf("Failed to listen on port %s | %v", portString, listenErr)
	}

	grpcListener := grpc.NewServer()                   // Creates server
	proto.RegisterNodeServer(grpcListener, nodeServer) // Registers our nodeServer object with the gRPC server

	log.Printf("Started listening on port %s", portString)

	serveListenerErr := grpcListener.Serve(listener) // Starts the server. THIS IS BLOCKING!
	if serveListenerErr != nil {
		log.Fatalf("Failed to serve listener | %v", serveListenerErr)
	}
}

// Bid Allows for gRPC bidding
func (nodeServer *NodeServer) Bid(ctx context.Context, bid *proto.AuctionBid) (*proto.BidAcknowledge, error) {
	if isPrimary { // If this node is primary
		if nodeServer.IsAuctionClosed() { // Bidding while auction is closed
			log.Printf("Bidder nr. %d, just tried to bid on a closed auction, what a nerd", bid.Id)
			return &proto.BidAcknowledge{Status: bidFail}, nil
		}

		return nodeServer.MakeBid(bid) // Auction isn't closed, make bid
	} else { // If this node isn't primary
		log.Printf("Echo bid {ID: %d, Bid: %d} from %d to %d", bid.Id, bid.Amount, listeningPort, echoPort)

		bidAcknowledge, err := echoNode.Bid(ctx, bid) // Echos bid to echoNode
		if err != nil {                               // Echoing failed, probably echoNode crashed
			log.Printf("Failed to echo bid on port %d", echoPort)

			ConnectToEchoNode(secondaryPort)                                          // Change echoNode to secondary node instead
			_, err := echoNode.PromoteSecondary(context.Background(), &proto.Empty{}) // Tell secondary to promote to primary since primary crashed
			if err != nil {                                                           // Secondary also down? Two crashes... so this won't happen :))))
				log.Fatalf("Failed to promote secondary node, this sucks :(")
			}

			return nodeServer.Bid(ctx, bid) // Re-attempts bid
		}

		log.Printf("Response {Status: %d}", bidAcknowledge.Status)
		return bidAcknowledge, nil
	}
}

// MakeBid Actually processes bid, and if it's a winning bid, it updates local auction data and replicates said data
func (nodeServer *NodeServer) MakeBid(bid *proto.AuctionBid) (*proto.BidAcknowledge, error) {
	log.Printf("Bidder nr. %d, just bid %d", bid.Id, bid.Amount)

	if bid.Amount > nodeServer.highestBid { // Incoming bid is higher
		log.Printf("New bid %d, is winning bid! (previous highest: %d)", bid.Amount, nodeServer.highestBid)
		nodeServer.highestBid = bid.Amount
		log.Printf("New auction leader: nr. %d (Previous leader: nr. %d)", bid.Id, nodeServer.highestBidder)
		nodeServer.highestBidder = bid.Id

		// We only replicate when data changes, which is when there's a new winning bid
		nodeServer.Replicate() // Replicate new data

		return &proto.BidAcknowledge{Status: bidSuccess}, nil // Bid success
	} else { // Incoming bid is lower
		log.Printf("New bid %d, is less than winning bid %d, ignoring...", bid.Amount, nodeServer.highestBid)

		return &proto.BidAcknowledge{Status: bidFail}, nil // Bid fail
	}
}

// Replicate Replicates auction data to secondary synchronously and to backups asynchronously
func (nodeServer *NodeServer) Replicate() {
	err := nodeServer.TryReplicate() // Attempt replication (can fail if secondary is down)
	if err != nil {                  // Secondary crashed
		ElectNewSecondary()    // Elect new secondary and update replicates
		nodeServer.Replicate() // Try again
	}
}

// TryReplicate Attempts to replicate auction state/data to secondary and backup nodes
func (nodeServer *NodeServer) TryReplicate() error {
	replicationData := &proto.AuctionData{ // Set up replication data package
		HighestBid:     nodeServer.highestBid,
		HighestBidder:  nodeServer.highestBidder,
		AuctionEndTime: timestamppb.New(nodeServer.auctionEndTime), // Convert Go time to protobuf Timestamp
	}

	// Replicate to secondary
	_, err := secondaryNode.ReplicateAuction(context.Background(), replicationData)
	if err != nil { // Secondary is down!
		return err // This is bad (return error so we can handle it in Replicate())
	}

	for _, backupNode := range backupNodes { // Loop over backup nodes for asynchronous replication
		go ReplicateToBackup(backupNode, replicationData) // Thrown to goroutine because we don't care about response
	}

	return nil
}

// ReplicateToBackup Replicates to a node without caring about any errors thrown. Meant to be used with goroutine
func ReplicateToBackup(backupNode proto.NodeClient, replicationData *proto.AuctionData) {
	_, backupRepErr := backupNode.ReplicateAuction(context.Background(), replicationData)
	if backupRepErr != nil {
		// We don't care
	}
}

// Result Allows for querying of auction state through gRPC
func (nodeServer *NodeServer) Result(ctx context.Context, empty *proto.Empty) (*proto.AuctionOutcome, error) {
	return &proto.AuctionOutcome{ // Returns auction state
			IsFinished: nodeServer.IsAuctionClosed(),
			HighestBid: nodeServer.highestBid,
			LeaderId:   nodeServer.highestBidder},
		nil
}

// IsAuctionClosed Checks if current time is after auctionEndTime
func (nodeServer *NodeServer) IsAuctionClosed() bool {
	return time.Now().After(nodeServer.auctionEndTime)
}

// ReplicateAuction Allows for a node to receive auction state data over gRPC. This simply overrides local data with received data
func (nodeServer *NodeServer) ReplicateAuction(ctx context.Context, data *proto.AuctionData) (*proto.Empty, error) {
	nodeServer.highestBid = data.HighestBid
	nodeServer.highestBidder = data.HighestBidder
	nodeServer.auctionEndTime = data.AuctionEndTime.AsTime() // Converts from protobuf time back to time.Time

	log.Printf("Replication data received {%d by %d, ends at %v}", nodeServer.highestBid, nodeServer.highestBidder, nodeServer.auctionEndTime.Format("15:04:05.9"))

	return &proto.Empty{}, nil
}

// PromoteSecondary Promotes a secondary node to primary. So this should only be called on the secondary node
func (nodeServer *NodeServer) PromoteSecondary(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	if isPrimary { // If we're already a primary, ignore call... (technically, this method gets called everytime a node realizes primary is down)
		return &proto.Empty{}, nil
	}

	log.Print("Promoting to primary...")
	isPrimary = true                  // Set local isPrimary state to true so gRPC calls are handled properly
	ElectNewSecondary()               // Updates secondary and backup nodes
	go nodeServer.CloseAuctionAfter() // Prints auction remaining time and also prints upon closing

	return &proto.Empty{}, nil
}

type logWriter struct {
}

// SetLogger Just some weird witchcraft to set custom print formatting, don't mind it
func SetLogger() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("15:04:05.9 ") + string(bytes))
}
