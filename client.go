package main

import (
	proto "Replica/gRPC"
	"bufio"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	ports []int
	activeNode
)

func main() {
	registerNodes()

	// also need to connect??

	// Listen to input: bid <amount>, result (query the auction)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input := scanner.Text()
		commands := strings.Fields(input) // This splits the input by white space

		if len(commands) == 0 {
			fmt.Println("Usage: bid <amount>, result")
			os.Exit(1)
		}

		if commands[0] == "bid" {
			amount, err := strconv.Atoi(commands[1])
			if err != nil {
				fmt.Println("Enter a valid amount")
			}
			bid(amount)
		} else if commands[1] == "result" {
			result()
		}
	}
}

func bid(amount int) {

}

func result() {

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
