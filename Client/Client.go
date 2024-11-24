package main

import (
	proto "Replica/gRPC"
	"bufio"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var idFlag = flag.Int("id", -1, "") // Simple flag for manually setting id

var (
	id              int32            // ID of the client
	ports           []int            // A slice/array of the ports of all available nodes
	activeNodeIndex = -1             // Index of the activeNode relative to the ports array
	activeNode      proto.NodeClient // The node we are currently sending our requests to
)

var (
	bidSuccess int32 = 0 // Constant signifying a bid was successful
	bidFail    int32 = 1 // Constant signifying a bid failed
	bidError   int32 = 2 // Constant signifying there was an error when trying to bid
)

func main() {
	SetLogger()     // Method for setting custom log.print() format
	ParseId()       // Sets ID using either flag or random number
	RegisterPorts() // Registers which nodes client knows

	Connect()        // Connects to available node
	ListenForInput() // Starts listening for user input
}

// ParseId Updates client ID to the idFlag if it exists, otherwise sets it to a random number
func ParseId() {
	flag.Parse()
	if *idFlag == -1 { // -1 is default value, therefore checks if it's unset
		id = int32(rand.Intn(100)) // random between 0-99
	} else {
		id = int32(*idFlag)
	}
	log.Printf("Client ID: %d", id)
}

// RegisterPorts Registers list of known node ports based on os.Args[]
func RegisterPorts() {
	// Parse each launch argument from string to (port) integer
	for _, parameter := range os.Args[1:] {
		port, err := strconv.Atoi(parameter)
		if err != nil {
			continue // We just ignore anything that isn't a valid port
		}
		ports = append(ports, port)
	}

	// Check if there are enough ports (minimum of 2 required for crash prevention)
	if len(ports) < 2 {
		log.Print("Usage: go run Client.go <port1> <port2> ... <portN>")
		os.Exit(1)
	}
}

// Connect Connects/sets active node to next available node
func Connect() {
	// Initially 0 (-1 + 1), but gets set to activeNodeIndex + 1,
	// so we always only loop over *new* nodes
	startIndex := activeNodeIndex + 1

	for portIndex, port := range ports[startIndex:] { // Loops over all known ports, starts at startIndex
		node, err := ConnectNode(port) // Attempts to connect to node at current port
		if err != nil {
			log.Printf("Failed to connect to node %d/%d with port %d | Error: %v", portIndex+startIndex+1, len(ports), port, err)
			continue // Could not connect to node, continue to next port
		}

		// If there are no errors, connect to first available node
		activeNode = node
		// Updates activeNodeIndex with correct index,
		// such that the next time Connect() is called,
		// loop starts at activeNodeIndex + 1
		activeNodeIndex = portIndex + startIndex
		log.Printf("Connected to node %d/%d with port %d", portIndex+startIndex+1, len(ports), port)
		break
	}
}

// ConnectNode Attempts to connect to node with given port
func ConnectNode(port int) (proto.NodeClient, error) {
	portString := fmt.Sprintf(":%d", 16000+port) // Format port string
	dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	connection, err := grpc.NewClient(portString, dialOptions) // Connect
	if err != nil {
		return nil, err // Return error if it fails to connect
	}

	return proto.NewNodeClient(connection), nil // Return node if it succeeds
}

// ListenForInput Constantly checks user input for either bid or result commands
func ListenForInput() {
	// Listen to input:
	//    bid <amount> (bids the given amount)
	//    result (queries the auction state)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input := strings.ToLower(scanner.Text())

		commands := strings.Fields(input) // This splits the input by white space

		if len(commands) == 0 { // Checks if input is empty
			log.Print("Usage: bid <amount>, result")
			continue
		}
		switch commands[0] { // Switches over possible commands
		case "bid":
			if len(commands) < 2 { // Checks for missing amount input
				log.Print("Please enter an amount to bid")
				continue
			}

			amount, err := strconv.Atoi(commands[1])
			if err != nil { // Inputted amount wasn't an integer
				log.Printf("Please enter a valid amount. %s, isn't valid", commands[1])
				continue
			}

			Bid(int32(amount)) // Actually makes bid
		case "result":
			Result()
		default: // Wasn't either of our options
			log.Print("Usage: bid <amount>, result")
		}
	}
}

// Bid Bids a given amount on the auction
func Bid(amount int32) {
	// Creates the proto bid message
	bidRequest := &proto.AuctionBid{
		Amount: amount,
		Id:     id,
	}

	// Call the Bid method in the activeNode
	ack, err := activeNode.Bid(context.Background(), bidRequest)
	if err != nil { // Active node has failed
		Connect()   // Switching to new node
		Bid(amount) // Re-try bid() now that active node has changed
		return
	}

	switch ack.Status { // Switch on bid result
	case bidSuccess:
		log.Printf("Bid of %d placed successfully!", amount)
	case bidFail:
		log.Print("Bid failed.")
	case bidError:
		log.Print("Error occurred during bidding.")
	default:
		log.Print("Unknown response status.")
	}
}

// Result Gets the current state of the auction
func Result() {
	// Call the Result method in the activeNode
	response, err := activeNode.Result(context.Background(), &proto.Empty{})
	if err != nil { // Active node has failed
		Connect() // Switching to new node
		Result()  // Re-try result() now that active node has changed
		return
	}

	// Print result depending on whether auction has finished
	if response.IsFinished {
		log.Printf("Auction finished, highest bid %d by nr. %d", response.HighestBid, response.LeaderId)
	} else {
		log.Printf("Auction ongoing, highest bid %d by nr. %d", response.HighestBid, response.LeaderId)
	}
}

type logWriter struct {
}

// SetLogger Just some weird witchcraft to set custom print formatting, don't mind it
func SetLogger() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format("15:04:05 ") + string(bytes))
}
