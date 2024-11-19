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

var (
	bidSuccess int32 = 0
	bidFail    int32 = 1
	bidError   int32 = 2
)

func main() {
	RegisterPorts()
	id = int32(rand.Intn(100))
	log.Printf("Client ID: %d", id)

	Connect()

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

func Connect() {
	for _, port := range ports {
		log.Printf("Port: %d", port)
		node, err := ConnectNode(port)
		if err != nil {
			log.Printf("Failed to connect to node with port %d: %v", port, err)
			continue // Could not connect to node, continue to next port
		}
		if activeNode == nil {
			activeNode = node
			break
		} else if node != activeNode {
			activeNode = node
			break
		}
	}
}

// ConnectNode connect to a Node from a port
func ConnectNode(port int) (proto.NodeClient, error) {
	portString := fmt.Sprintf(":1600%d", port) // Format port string
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
	}

	// Handle the response
	if response.IsAuctionFinished {
		log.Printf("The auction is finished, highest bid was %d by bidder nr. %d", response.HighestBid, response.WinningBidder)
	} else {
		log.Printf("The auction is still ongoing. The highest bid is %d by bidder nr. %d", response.HighestBid, response.WinningBidder)
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
