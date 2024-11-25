# Replica

To run the program:
1. Clone the repository to your own machine.
2. Open the terminal and navigate to the project directory.
3. In the terminal, go to the ``Client/`` folder, and begin by running the clients: ``go run Client.go <port1> <port2> .... <portN>``, where the ports are the list of known nodes for said client.
4. To run a node, go to the ``Node/`` folder, then run the commands depending on the type of node you want to start:
   - Primary node: ``go run Node.go -p=true <currentPort> <secondaryPort> <port1> ... <portN>``. In ``<port1> ... <portN>`` include ports of all backup nodes.
   - Secondary node: ``go run Node.go <currentPort> <primaryPort> <secondaryPort> <port1> ... <portN>``
   - Normal node: ``go run Node.go <currentPort> <primaryPort> <secondaryPort>``
   - IMPORTANT! The order in which nodes are run must be as follows: Backup nodes > Secondary node > Primary node.
   - Explanation of port names:
      - ``currentPort`` is the port this node listens on.
      - ``primaryPort`` is the port belonging to the primary node (primary's ``currentPort``).
      - ``secondaryPort`` is the port belonging to the secondary node (secondary's ``currentPort``).
      - ``port[X]`` is the port of any node that isn't the primary. The number of ports with numbers listed, is the minimum required amount (``<port1> <port2> ... <portN>`` means at minimum 2 ports are required).
7. With the client you can run the following commands:
- `` bid <amount>``, where amount is a number. This will attempt to place a bid, and depending on the current highest bid, validate or not validate the bid.
- ``result``, which will return the result of auction.
7. To crash any of the nodes, press ``CTRL + C`` or manually shut down the terminal.
8. Close the program by shutting down all terminals.
