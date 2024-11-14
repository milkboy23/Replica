package main

import (
	"fmt"
	"os"
	"strconv"
)

var (
	ports []int
)

func main() {
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
