# Replica

To run the program:
1. Clone the repository to your own machine.
2. Open the terminal and navigate to the project directory
3. In the terminal, begin by running the clients: go run client.go <port1> <port2> .... <portN> | where each of the ports are the list of known nodes for the client
4. To run a node go to the node folder then run the commands depending on the type of node you want to start:
   - Primary node: ``go run node.go <currentPort> <secondaryPort> <port1> ... <portN> -p=true``
   - Secondary node: ``go run node.go <currentPort> <primaryPort> <secondaryPort/currentPort> <port1> ... <portN>``
   - Normal node: ``go run node.go <currentPort> <primaryPort> <secondaryPort>``
IMPORTANT! The order in which nodes are run must be as follows: Backup nodes, Secondary node, Primary node.
6. With the client you can run the following commands:
- bid ``<amount>`` | where amount is a number. This will attempt to place a bid, and depending on the current highest bid, validate or not validate the bid.
- result | which will return the result of auction.
7. To crash any of the nodes, type ctrl + c or shut down the terminal.
8. Close the program by shutting down all terminals.
