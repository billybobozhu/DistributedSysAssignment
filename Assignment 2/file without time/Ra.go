package main

import (
	"bufio"
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
var localClock int
var inCritical bool
var waIting bool
var received_all_replies bool
var sharedResource *net.UDPConn
var queued_request []int
var my_logicClock int
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
	if my_logicClock < lc_pj {
		//I am priority =)
		return true
	} else if lc_pj > my_logicClock {
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
	str_lc := strconv.Itoa(localClock)
	//transformar id em string
	str_id := strconv.Itoa(id)
	// concatenar todas
	mymsg := str_id + "," + str_lc + ",reply"
	buf := []byte(mymsg)
	//Enviar mensagem para o pj
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

func procurar_in_list(pj_id int) bool {

	for _, i := range replies_received {
		if i == pj_id {
			return true
		}
	}
	return false
}

func doServerJob() {
	//Ler (uma vez somente) da conexão UDP a mensagem
	//Escreve na tela a msg recebida

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
				fmt.Printf("\nThreaded %d with clock %d, because I am  = %t, I am waiting = %t, my id = %d, my clock = %d\n", pj_id, lc_pj, inCritical, waIting, id, my_logicClock)
				queue_request_from_pj(pj_id)

			} else {
				//reply immediately to pj
				fmt.Printf("\nsending reply <id,clock> = < %d , %d > for %d \n", id, localClock, pj_id)
				reply2pj(pj_id)
			}
		} else if stream_msg[2] == "reply" {
			//If menssage is reply
			//procurar pj na lista de replies recebidas
			if !procurar_in_list(pj_id) {
				//se nao tiver coloca na lista de replies recebidas
				replies_received = append(replies_received, pj_id)
			}
			//Se recebeu todas replies
			if len(replies_received) >= nServers {
				//ativar flag de recebido todas replies
				fmt.Println("I received all replies: ", len(replies_received))
				received_all_replies = true
			}
		} else {
			fmt.Println("Unknown message ", aux)
		}
		//update logical clock = max(my,other) + 1
		localClock = MaxInt(localClock, lc_pj) + 1
		fmt.Println("I updated my watch to ", localClock)
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
	localClock = 0
}

func readInput(ch chan string) {
	// Non-blocking async routine to listen for terminal input
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

func use_CS(my_logicClock int, text_simples string) {
	//Send message and sleep
	inCritical = true
	str_my_logicClock := strconv.Itoa(my_logicClock)
	str_id := strconv.Itoa(id)
	mymsg := str_id + "," + str_my_logicClock + "," + text_simples
	//message -->byte
	buf := []byte(mymsg)
	// add to survilliance
	_, err := sharedResource.Write(buf)
	if err != nil {
		fmt.Println(mymsg, err)
	}

}

func Requesting_access_CS(lclock int) {
	//Build menssage
	str_logical_clock := strconv.Itoa(lclock)

	str_id := strconv.Itoa(id)
	mymsg := str_id + "," + str_logical_clock + ",request"
	buf := []byte(mymsg)
	for _, conn2process := range CliConn {
		//send to all the clients
		_, err := conn2process.Write(buf)
		if err != nil {
			fmt.Println(mymsg, err)
		}
	}
}

func reply_any_queued_request() {

	//Build menssage
	str_logical_clock := strconv.Itoa(localClock)
	//transformar id to string
	str_id := strconv.Itoa(id)
	mymsg := str_id + "," + str_logical_clock + ",reply"
	//transformar mensagem em bytes
	buf := []byte(mymsg)
	//Reply to all queued processes
	for _, id := range queued_request {

		index := id - 1
		//reply queued request
		_, err := CliConn[index].Write(buf)
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

func Ricart_Agrawala(my_logicClock int, text_simples string) {

	waIting = true
	Requesting_access_CS(my_logicClock)
	//Wait until received N-1 replies
	fmt.Println("Waiting for all replies")
	for !received_all_replies {
	}
	fmt.Println("entrer CS!")
	use_CS(my_logicClock, text_simples)
	fmt.Println("in CS!")
	time.Sleep(time.Second * 10)

	fmt.Println("exit  CS!")
	release_CS()
	fmt.Println("Free CS!")
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
	ch := make(chan string)
	go readInput(ch)
	for {

		//Server
		go doServerJob()
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
						fmt.Printf("Requesting access with ID = %d and Logical Clock = %d\n", id, localClock)
						text_simples := "CS here"
						my_logicClock = localClock
						go Ricart_Agrawala(my_logicClock, text_simples)
					}

				} else {

					localClock++
					fmt.Printf("Updated logicalClock for %d \n", localClock)
				}
			} else {

				fmt.Println("Channel closed!")

			}

		default:

			// Do nothing in the non-blocking approach.

			time.Sleep(time.Second * 1)
		}

		// Wait a while
		time.Sleep(time.Second * 1)
	}
}
