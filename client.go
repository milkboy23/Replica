package main

import (
	proto "Replica/gRPC"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
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
	registerNodes()
	id = int32(rand.Intn(100))

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
			bid(int32(amount))
		} else if commands[1] == "result" {
			result()
		}
	}
}

func bid(amount int32) {
	bidRequest := &proto.Bid{
		Amount: amount,
		Id:     id,
	}

	// Call the Bid method in the activeNode
	ack, err := activeNode.Bid(context.Background(), bidRequest)
	if err != nil {
		// node failed handle by switching to new node
	}

	switch ack.Status {
	case 0:
		fmt.Printf("Bid of %s placed successfully! \n", amount)
	case 1:
		fmt.Println("Bid failed.")
	case 2:
		fmt.Println("Error occurred during bidding.") // maybe reconnect
	default:
		fmt.Println("Unknown response status.") // maybe reconnect
	}
}

func result() {
	// Call the Result method in the activeNode
	response, err := activeNode.Result(context.Background(), &proto.Empty{})
	if err != nil {
		// node failed handle by switching to new node
	}

	// Handle the response
	if response.IsAuctionFinished {
		fmt.Printf("The auction is finished. The highest bid was %d.\n", response.HighestBid)
	} else {
		fmt.Printf("The auction is still ongoing. The highest bid is %d.\n", response.HighestBid)
	}
}

// Connect to a Node from a port
func ConnectNode(port int) proto.NodeClient {
	portString := fmt.Sprintf(":%d", port) // Format port string
	dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	connection, connectionEstablishErr := grpc.NewClient(portString, dialOptions)
	if connectionEstablishErr != nil {
		log.Fatalf("Could not establish connection on port %s | %v", portString, connectionEstablishErr)
	}

	return proto.NewNodeClient(connection)
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
			os.Exit(1)
		}
		ports = append(ports, port)
	}

	// Print each port in the list?
	for _, port := range ports {
		fmt.Printf("Port: %d\n", port)
	}
}
