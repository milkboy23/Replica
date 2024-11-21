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

var idFlag = flag.Int("id", -1, "")

var (
	id              int32
	ports           []int
	activeNodeIndex int
	activeNode      proto.NodeClient
)

var (
	bidSuccess int32 = 0
	bidFail    int32 = 1
	bidError   int32 = 2
)

func main() {
	SetLogger()
	ParseId()
	RegisterPorts()

	Connect()
	ListenForInput()
}

func ParseId() {
	flag.Parse()
	if *idFlag == -1 {
		id = int32(rand.Intn(100))
	} else {
		id = int32(*idFlag)
	}
	log.Printf("Client ID: %d", id)
}

// RegisterPorts register ports from input
func RegisterPorts() {
	// Parse each port
	for _, parameter := range os.Args[1:] {
		port, err := strconv.Atoi(parameter)
		if err != nil {
			continue
		}
		ports = append(ports, port)
	}

	// Check if there are ports
	if len(ports) < 2 {
		log.Print("Usage: go run Client.go <port1> <port2> ... <portN>")
		os.Exit(1)
	}
}

func Connect() {
	var portIndex = 0
	for _, port := range ports {
		portIndex++
		node, err := ConnectNode(port)
		if err != nil {
			log.Printf("Failed to connect to node with port %d/%d on %d: %v", portIndex, len(ports), port, err)
			continue // Could not connect to node, continue to next port
		}
		if activeNode == nil || portIndex > activeNodeIndex {
			activeNode = node
			activeNodeIndex = portIndex
			log.Printf("Connected to node %d/%d with port %d", portIndex, len(ports), port)
			break
		}
	}
}

// ConnectNode connect to a Node from a port
func ConnectNode(port int) (proto.NodeClient, error) {
	portString := fmt.Sprintf(":%d", 16000+port) // Format port string
	dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	connection, err := grpc.NewClient(portString, dialOptions)
	if err != nil {
		return nil, err // Return error if it fails to connect
	}

	return proto.NewNodeClient(connection), nil // Return Node if it succeeds
}

func ListenForInput() {
	// Listen to input: bid <amount>, result (query the auction)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input := strings.ToLower(scanner.Text())

		commands := strings.Fields(input) // This splits the input by white space

		if len(commands) == 0 {
			log.Print("Usage: bid <amount>, result")
			continue
		}
		switch commands[0] {
		case "bid":
			if len(commands) < 2 {
				log.Print("Please enter an amount to bid")
				continue
			}
			amount, err := strconv.Atoi(commands[1])
			if err != nil {
				log.Printf("Please enter a valid amount. %s, isn't valid", commands[1])
				continue
			}
			Bid(int32(amount))
		case "result":
			Result()
		default:
			log.Print("Usage: bid <amount>, result")
		}
	}
}

// Bid in the auction an amount
func Bid(amount int32) {
	bidRequest := &proto.AuctionBid{
		Amount: amount,
		Id:     id,
	}

	// Call the Bid method in the activeNode
	ack, err := activeNode.Bid(context.Background(), bidRequest)
	if err != nil {
		Connect() // node failed handle by switching to new node
		Bid(amount)
		return
	}

	switch ack.Status {
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

// Result get the auction result
func Result() {
	// Call the Result method in the activeNode
	response, err := activeNode.Result(context.Background(), &proto.Empty{})
	if err != nil {
		Connect() // node failed handle by switching to new node
		Result()
		return
	}

	// Handle the response
	if response.IsFinished {
		log.Printf("Auction finished, highest bid %d by nr. %d", response.HighestBid, response.LeaderId)
	} else {
		log.Printf("Auction ongoing, highest bid %d by nr. %d", response.HighestBid, response.LeaderId)
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
