start cmd /k go run replicaServer.go 1
start cmd /k go run CentralizedServer.go 1
start cmd /k go run centralizedClientTime.go 0 :10001 0 :5000 write


Here are example bat commands

start cmd /k go run replicaServer.go 1  --> RUN SERVER 1 IS HOW MANY CLIENTS ARE CONNECTING.
go run xxx.go <no of clients>

start cmd /k go run centralizedClientTime.go 0 :10001 0 :5000 write 
--> the first parameter is client id , then port number , then page the client wants to operate,
then port number listening for response, then operation

exmaple：

go run xxx.go <client id> <port number> <page> <port number listening> <operation>