start cmd /k go run centralizedServer.go 5
start cmd /k go run centralizedClientTime.go 0 :10001 1 :5000 read
start cmd /k go run centralizedClientTime.go 1 :10002 2 :5001 read
start cmd /k go run centralizedClientTime.go 2 :10002 3 :5002 read
start cmd /k go run centralizedClientTime.go 3 :10002 4 :5003 read
start cmd /k go run centralizedClientTime.go 4 :10002 0 :5004 read






