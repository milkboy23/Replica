# Replica

To run the program:
1. Clone the repository to your own machine.
2. Open the terminal and navigate to the project directory
3. To run a node go to the node folder then run the commands depending on the type of node you want to start:
   - Primary node: ``go run node.go <currentPort> <secondaryPort> <port1> ... <portN> -p=true``
   - Secondary node: ``go run node.go <currentPort> <primaryPort> <secondaryPort/currentPort> <port1> ... <portN>``
   - Normal node: ``go run node.go <currentPort> <primaryPort> <secondaryPort>``
5. To run the client navigate to the client folder and run: go run client.go <port1> <port2> .... <portN> | where each of the ports are the list of known nodes for the client

With the client you can run the following commands:
- bid <amount> | where amount is a number. this will place a bid and if it is first run then it will register the client as a bidder
- result | which will return result of auction


6. To close the program ctrl + c or shut down the terminals.
