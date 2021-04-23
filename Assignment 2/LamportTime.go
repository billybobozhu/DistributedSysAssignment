package main

import (
	"fmt"
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

func CheckError(err1 error) {
	if err1 != nil {
		fmt.Println("Erro: ", err1)
		os.Exit(0)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}

func Am_I_priority(pj_id int, lc_pj int) bool {

	//my logical clock at request is lower than another's
	if lc_requisition < lc_pj {
		//I am priority =)
		return true
	} else if lc_pj > lc_requisition {
		//my logical clock at request isn't lower than another's
		//I'm not priority TT
		return false
	} else {
		//my logical clock at request is equal to another's
		if id < pj_id {
			//but my id is lower
			//So i am priority =)
			return true
		} else {
			//my id isn't lower
			//I'm not priority TT
			return false
		}
	}
}

func queue_request_from_pj(pj_id int) {
	//queue request from pj without replying
	queued_request = append(queued_request, pj_id)
}

func reply2pj(pj_id int) {
	//reply immediately to pj
	//build reply menssage
	str_lc := strconv.Itoa(mylogicalClock)
	//transformar id em string
	str_id := strconv.Itoa(id)
	mymsg := str_id + "," + str_lc + ",reply"
	buf := []byte(mymsg)
	index := pj_id - 1
	//reply to pj
	_, err := CliConn[index].Write(buf)
	if err != nil {
		fmt.Println(mymsg, err)
	}
}

func MaxInt(x1, x2 int) int {

	if x1 > x2 {
		return x1
	}
	return x2
}

func search_in_list(pj_id int) bool {

	for _, i := range replies_received {
		if i == pj_id {
			return true
		}
	}
	return false
}

func doServerJob() {

	buf := make([]byte, 1024)
	for {

		n, _, err := ServConn.ReadFromUDP(buf)
		//aux = "id,logical_clock"
		aux := string(buf[0:n])
		//stream_msg = ["id" , "logical_clock"]
		stream_msg := strings.Split(aux, ",")
		str_pj_id := stream_msg[0]
		str_lc_pj := stream_msg[1]
		pj_id, err := strconv.Atoi(str_pj_id)
		lc_pj, err := strconv.Atoi(str_lc_pj)
		fmt.Println("I received message from ", str_pj_id, " with clock ", str_lc_pj, " message type ", stream_msg[2])
		//If menssage is request
		if stream_msg[2] == "request" {
			//I am in CS or (I want CS and preference is mine) then queue pj
			if inCritical || (waIting && Am_I_priority(pj_id, lc_pj)) {
				//queue request from pj withou replying
				fmt.Printf("\nThreaded %d with clock %d, because I am  = %t, I am waiting = %t, my id = %d, my clock = %d\n", pj_id, lc_pj, inCritical, waIting, id, lc_requisition)
				queue_request_from_pj(pj_id)
				fmt.Println(queued_request)

			} else {
				//reply immediately to pj
				queue_request_from_pj(pj_id)
				fmt.Printf("\nsending reply <id,clock> = < %d , %d > for %d \n", id, mylogicalClock, pj_id)
				fmt.Println("my request queue is:", queued_request)
				reply2pj(pj_id)
			}
		} else if stream_msg[2] == "reply" {
			//If message is reply
			//search for pj in the list of received replies
			if !search_in_list(pj_id) {
				//if you don't have a list of replies received
				replies_received = append(replies_received, pj_id)
			}
			//If you received all replies
			if len(replies_received) >= nServers {
				fmt.Println("I received all replies: ", len(replies_received))
				received_all_replies = true
			}
		} else if stream_msg[2] == "release" {
			fmt.Println("I received release message : ", pj_id)
			queued_request = queued_request[1:]

		} else {
			// fmt.Println("Unknown message ", aux)
		}
		//update logical clock = max(my,other) + 1
		mylogicalClock = MaxInt(mylogicalClock, lc_pj) + 1
		fmt.Println("I updated my watch to ", mylogicalClock)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func initConnections() {
	id, _ = strconv.Atoi(os.Args[1])
	myPort = os.Args[id+1]
	nServers = len(os.Args) - 2

	connections := make([]*net.UDPConn, nServers, nServers)

	for i := 0; i < nServers; i++ {

		port := os.Args[i+2]

		ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1"+string(port))
		PrintError(err)

		LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
		PrintError(err)

		connections[i], err = net.DialUDP("udp", LocalAddr, ServerAddr)
		PrintError(err)

	}
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	PrintError(err)
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	PrintError(err)
	sharedResource, err = net.DialUDP("udp", LocalAddr, ServerAddr)
	PrintError(err)

	CliConn = connections

	/* Lets prepare a address at any address at port 10001*/
	ServerAddr, err = net.ResolveUDPAddr("udp", myPort)
	CheckError(err)

	/* Now listen at selected port */
	ServConn, err = net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	//init process's logical clock with 0
	mylogicalClock = 0
}

// func readInput(ch chan string) {
// 	// Non-blocking async routine to listen for terminal input
// 	reader := bufio.NewReader(os.Stdin)
// 	for {
// 		text, _, _ := reader.ReadLine()
// 		ch <- string(text)
// 	}
// }

func use_CS(lc_requisition int, text_simples string) {

	str_lc_requisition := strconv.Itoa(lc_requisition)
	str_id := strconv.Itoa(id)
	mymsg := str_id + "," + str_lc_requisition + "," + text_simples
	buf := []byte(mymsg)
	_, err := sharedResource.Write(buf)
	if err != nil {
		fmt.Println(mymsg, err)
	}

}

func Requesting_access_CS(id int, lclock int) {
	//Build request message
	str_logical_clock := strconv.Itoa(lclock)
	str_id := strconv.Itoa(id)
	mymsg := str_id + "," + str_logical_clock + ",request"
	buf := []byte(mymsg)
	//send request to all the others
	for _, conn2process := range CliConn {

		_, err := conn2process.Write(buf)
		if err != nil {
			fmt.Println(mymsg, err)
		}
	}
}

func reply_any_queued_request() {

	// build reply message
	str_logical_clock := strconv.Itoa(mylogicalClock)
	//transformar id em string
	str_id := strconv.Itoa(id)
	mymsg := str_id + "," + str_logical_clock + ",release"

	buf := []byte(mymsg)
	//Reply to all queued processes
	for _, conn2process := range CliConn {

		_, err := conn2process.Write(buf)
		if err != nil {
			fmt.Println(mymsg, err)
		}
	}
}

func release_CS() {
	//to exit the CS
	//released, reset the flags.
	inCritical = false
	waIting = false
	received_all_replies = false
	//reply to any queued request
	reply_any_queued_request()
	//clear reply received list
	replies_received = nil

}

func lamport(id int, lc_requisition int, text_simples string) {

	// waIting = true
	now := time.Now()
	nsec := now.UnixNano()
	fmt.Println(nsec)
	Requesting_access_CS(id, lc_requisition)
	fmt.Println("Check if is the earlist in the queue")
	fmt.Println("here is the queue:", queued_request)
	time.Sleep(time.Second * 10)
	for queued_request[0] != id || !received_all_replies {
	}
	fmt.Println("entrer CS!")
	use_CS(lc_requisition, text_simples)
	fmt.Println("here is the queue:", queued_request)
	fmt.Println("in CS!")
	time.Sleep(time.Second * 10)
	fmt.Println("here is the queue:", queued_request)

	fmt.Println("exit  CS!")
	fmt.Println("here is the queue:", queued_request)
	release_CS()
	fmt.Println("Free CS!")
	now1 := time.Now()
	nsec1 := now1.UnixNano()
	fmt.Println(nsec1)
	fmt.Println("here is the queue:", queued_request)
}

func main() {
	initConnections()
	inCritical = false
	waIting = false
	//The closing of connections should stay here, so it only closes
	//connection when the main dies
	defer ServConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}
	//Todo Process will do the same thing: listen to msg and send infinites
	//i’s para os outros processos
	//ch := make(chan string)
	//go readInput(ch)
	for {

		//Server
		go doServerJob()
		// When there is a request (from stdin). Do it!
		// select {
		// case x, valid := <-ch:
		// 	if valid {
		// 		compare, _ := strconv.Atoi(x)
		// 		if compare != id && x == "x" {
		// 			//See if you are in CS or waiting
		// 			if inCritical || waIting {
		// 				fmt.Println("x ignored!")
		// 			} else {
		// 				fmt.Printf("Requesting access with ID = %d and Logical Clock = %d\n", id, mylogicalClock)
		// 				text_simples := "CS here"
		// 				lc_requisition = mylogicalClock
		// 				go lamport(id, lc_requisition, text_simples)
		// 			}

		// 		} else {

		// 			mylogicalClock++
		// 			fmt.Printf("Updated logicalClock for %d \n", mylogicalClock)
		// 		}
		// 	} else {

		// 		fmt.Println("Channel closed!")

		// 	}

		// default:

		// 	// Do nothing in the non-blocking approach.

		// 	time.Sleep(time.Second * 1)
		// }
		// time.Sleep(time.Second * 2)
		time.Sleep(time.Second * 3)
		fmt.Printf("Requesting access with ID = %d and Logical Clock = %d\n", id, mylogicalClock)
		text_simples := "CS here"
		lc_requisition = mylogicalClock
		go lamport(id, lc_requisition, text_simples)
		// Wait a while
		time.Sleep(time.Second * 10000)
	}
}
