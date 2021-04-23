package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

var id int
var message string
var condition bool

const connectionNum int = 10

func main() {
	// go serverJob()\

	var queued_request []int
	var connectionPool = make([]net.Addr, connectionNum)
	pc, err := net.ListenPacket("udp", ":6000")
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	buffer := make([]byte, 3)
	fmt.Println("Waiting for client...")
	for {

		_, addr, err := pc.ReadFrom(buffer)
		fmt.Println("incoming address", addr)
		if err == nil {

			rcvMsq := string(buffer)
			fmt.Println("Received: " + rcvMsq)
			s := strings.Split(rcvMsq, ",")
			newid, _ := strconv.Atoi(s[0])
			newmsg, _ := strconv.Atoi(s[1])

			connectionPool[newid] = addr
			fmt.Println("connected client : ", connectionPool)

			if newmsg == 1 { //request
				fmt.Println("message received with code number 1 : request , request from: client ", newid)
				queued_request = append(queued_request, newid)
				fmt.Println("queue appended! now queue is :", queued_request)

			} else if newmsg == 2 { //release
				fmt.Println("lock is released")
				queued_request = queued_request[1:]
				condition = false

			}

			if condition == false && len(queued_request) != 0 {
				mymsg := strconv.Itoa(queued_request[0]) + "," + "1" // 1= allow
				if _, err := pc.WriteTo([]byte(mymsg), connectionPool[queued_request[0]]); err != nil {
					fmt.Println("error on write: " + err.Error())
				}
				condition = true

			}

		} else {
			fmt.Println("error: " + err.Error())
		}
	}

}
