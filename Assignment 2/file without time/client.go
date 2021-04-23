package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var err string
var myPort string
var nServers int
var CliConn []*net.UDPConn
var ServConn *net.UDPConn
var id int
var mylogicalClock int
var inCritical bool
var waIting bool
var received_all_replies bool
var sharedResource *net.UDPConn
var queued_request []int
var lc_requisition int
var replies_received []int

func readInput(ch chan string) {
	// Non-blocking async routine to listen for terminal input
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

func server(conn *net.UDPConn) {
	buffer := make([]byte, 3)
	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("lock Server : ", addr)
	fmt.Println("Received from lock server : ", string(buffer[:n]))
	k := string(buffer[:n])
	s := strings.Split(k, ",")

	newid, _ := strconv.Atoi(s[0])
	reply, _ := strconv.Atoi(s[1])
	if reply == 1 { // enter cs allowed
		fmt.Printf("client %d ENTER CS !\n", newid)
		time.Sleep(time.Second * 3)
		fmt.Printf("client %d QUIT CS !\n", newid)

		strId := strconv.Itoa(newid)
		strMessage := strconv.Itoa(2)
		//E concatenar todas
		mymsg := strId + "," + strMessage

		buf := []byte(mymsg)

		_, err = conn.Write(buf)
		if err != nil {
			fmt.Println(mymsg, err)
		}

	}

}
func main() {
	inCritical = false
	waIting = false
	hostName := "localhost"
	portNum := "6000"
	id, _ = strconv.Atoi(os.Args[1])
	myPort = os.Args[2]

	service := hostName + ":" + portNum
	service2 := hostName + ":" + myPort

	RemoteAddr, err := net.ResolveUDPAddr("udp", service)
	LocalAddr, err := net.ResolveUDPAddr("udp", service2)
	conn, err := net.DialUDP("udp", LocalAddr, RemoteAddr)

	// note : you can use net.ResolveUDPAddr for LocalAddr as well
	//        for this tutorial simplicity sake, we will just use nil

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Established connection to %s \n", service)
	log.Printf("Remote UDP address : %s \n", conn.RemoteAddr().String())
	log.Printf("Local UDP client address : %s \n", conn.LocalAddr().String())

	defer conn.Close()
	ch := make(chan string)
	go readInput(ch)
	// write a message to server
	for {
		go server(conn)
		//Server

		// When there is a request
		select {
		case x, valid := <-ch:
			if valid {
				compare, _ := strconv.Atoi(x)
				if compare != id && x == "x" {
					//See if you are in CS or waiting
					if inCritical || waIting {
						fmt.Println("x ignored!")
					} else {
						fmt.Printf("Requesting access with ID = %d \n", id)
						// text_simples := "CS here"

						strId := strconv.Itoa(id)
						strMessage := strconv.Itoa(1)
						//E concatenar todas
						mymsg := strId + "," + strMessage

						buf := []byte(mymsg)

						_, err = conn.Write(buf)
						if err != nil {
							fmt.Println(mymsg, err)
						}
					}

				} else {
					fmt.Println("not requesting for the lock, just for fun")

				}
			} else {

				fmt.Println("Channel closed!")

			}

		default:

			// Do nothing in the non-blocking approach.

			time.Sleep(time.Second * 1)
		}

		// Wait a while
		// time.Sleep(time.Second * 1)
	}

	// message := []byte("")

	// _, err = conn.Write(message)

	// if err != nil {
	// 	log.Println(err)
	// }

	// // receive message from server
	// buffer := make([]byte, 1024)
	// n, addr, err := conn.ReadFromUDP(buffer)

	// fmt.Println("lock Server : ", addr)
	// fmt.Println("Received from lock server : ", string(buffer[:n]))

	readin()
}

func readin() {
	fmt.Scanln()
}
