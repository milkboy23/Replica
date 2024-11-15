package main

import (
	proto "Replica/gRPC"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

var (
	id         int32
	ports      []int
	activeNode proto.NodeClient
)

func main() {
	RegisterPorts()
	id = int32(rand.Intn(100))

	Connect()

	// Listen to input: bid <amount>, result (query the auction)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input := strings.ToLower(scanner.Text())

		commands := strings.Fields(input) // This splits the input by white space

		if len(commands) == 0 {
			fmt.Println("Usage: bid <amount>, result")
		}

		if commands[0] == "bid" {
			amount, err := strconv.Atoi(commands[1])
			if err != nil {
				fmt.Println("Enter a valid amount")
				continue
			}
			Bid(int32(amount))
		} else if commands[1] == "result" {
			Result()
		}
	}
}

func Connect() {
	for _, port := range ports {
		fmt.Printf("Port: %d\n", port)
		node, err := ConnectNode(port)
		if err != nil {
			continue // Could not connect to node, continue to next port
		}
		activeNode = node
	}
}

// ConnectNode connect to a Node from a port
func ConnectNode(port int) (proto.NodeClient, error) {
	portString := fmt.Sprintf(":%d", port) // Format port string
	dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	connection, err := grpc.NewClient(portString, dialOptions)
	if err != nil {
		return nil, err // Return error if it fails to connect
	}

	return proto.NewNodeClient(connection), nil // Return Node if it succeeds
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
	}

	switch ack.Status {
	case 0:
		fmt.Printf("Bid of %s placed successfully! \n", amount)
	case 1:
		fmt.Println("Bid failed.")
	case 2:
		fmt.Println("Error occurred during bidding.")
	default:
		fmt.Println("Unknown response status.")
	}
}

// Result get the auction result
func Result() {
	// Call the Result method in the activeNode
	response, err := activeNode.Result(context.Background(), &proto.Empty{})
	if err != nil {
		Connect() // node failed handle by switching to new node
	}

	// Handle the response
	if response.IsAuctionFinished {
		fmt.Printf("The auction is finished. The highest bid was %d.\n", response.HighestBid)
	} else {
		fmt.Printf("The auction is still ongoing. The highest bid is %d.\n", response.HighestBid)
	}
}

// RegisterPorts register ports from input
func RegisterPorts() {
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
			os.Exit(1)
		}
		ports = append(ports, port)
	}
}
