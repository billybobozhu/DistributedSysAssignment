IN COLLABORATION WITH LUO YIFAN 

Compare the performance of the three implementations (in terms of time). In
particular, increase the number of nodes simultaneously requesting to enter the critical
sections to investigate the performance trend for each of the protocol. For each experiment
and for each protocol, compute the time between the first request and all the requesters exit
the critical section. If you have ten nodes, therefore, your performance table should contain a
total of 30 entries (ten entries for each of the three protocols). 

IF we set time for critical session is 10sec plus 10sec assumed delay in UDP  we can observed that : 
(refer to compare_when cs is long )

With increase the number of nodes, all three protocols show a linearly increasing trend.
Generally centralized lock server perform better when there are many nodes and RA protocol perform better with fewer nodes.
Lamport queue protocol has the poorest performance among the three protocols.
However, the difference between these three protocols are very small (in millisecond scale) when the node number is small.

If time for cs is 0 

With increase the number of nodes, RA and lamport queue shows a nonlinear increase and centralized server still perform a linearly increase
However, lamport still has poorest performance and centralized server perform better when node number is larger.

If you want to modify the cs time
please go to the file and comment the line  time.sleep()

How to run the code?

For RA and Lamport: go run <filename> <clientID> <port number 1> <port number 2>...

for example: if there we want three nodes we start three terminals and run (or use bat):
go run RaTime.go 1 :1000 :1001 :1002
go run RaTime.go 2 :1000 :1001 :1002
go run RaTime.go 3 :1000 :1001 :1002



For centralized server: go run <filename> <clientID> <port number>

for example: if there we want three nodes we start three terminals and run (or use bat)
go run CentralizedServer.go
go run CentralizedClientTime.go 1 :1000
go run CentralizedClientTime.go 2 :1001
go run CentralizedClientTime.go 3 :1002