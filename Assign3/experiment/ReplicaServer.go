package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var id int
var message string
var condition bool
var activate bool

type page struct {
	pageNumber int
	copySet    []int
	owner      int
}

const connectionNum int = 10

var me = Client{}
var clientsInfo = make(map[int]*Client)
var serverCons = make(map[int]net.Conn)
var coordinator int
var leader bool

type Client struct {
	s_conn net.Conn
	pid    int
	port   int
	leader bool
}

//Directory for Bully Algorithm

func readFromServer(conn net.Conn, ch chan<- string, blockCh chan bool, connectCh chan bool) {
	reader := bufio.NewReader(conn)

	for {

		message, err := reader.ReadString('\n')
		//conn.Read(recvdSlice)
		if message != "OK\n" {
			if err != nil {
				break
			}
			s := strings.Split(message, ":")
			pid, _ := strconv.Atoi(s[0])
			port, _ := strconv.Atoi(s[1])
			leader, _ := strconv.ParseBool(s[2][:len(s[2])-1])
			if leader == true {
				coordinator = pid
				fmt.Println("coordinator id  : ", coordinator)

			}
			if (Client{}) == me {
				me.port = port
				me.pid = pid
				me.leader = leader
				blockCh <- true

			} else {
				clientsInfo[pid] = &Client{port: port, leader: leader, pid: pid}

			}
			//	ch <- fmt.Sprintf("%s", string(recvdSlice))

		} else {
			fmt.Println("OK")
			connectOthers()
		}

	}

}

func printRoutine(serverChan chan string) {
	for {
		select {
		case msg := <-serverChan:
			{
				fmt.Println(msg)
			}
		}
	}
}
func serve(ch chan bool, clientCh chan string, rmCh chan int, allCh chan string) {
	_ = <-ch
	fmt.Println("my Process id : ", me.pid)

	fmt.Println("Listening at port ", strconv.Itoa(me.port))

	listen, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(me.port))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listen.Accept()

		if err != nil {
			fmt.Println("Error while accepting a new connection", err)
			//		continue //move to the next iteration
		}

		go handleConnection(conn, clientCh, rmCh, allCh)

	}
}
func getleaderId() int {
	max := me.pid
	for id, _ := range clientsInfo {
		if id > max {
			max = id
		}
	}
	return max
}
func randomDetector() int {
	slice := make([]int, len(clientsInfo))
	itr := 0
	for pid, _ := range clientsInfo {
		slice[itr] = pid
		itr += 1
	}
	ind := slice[rand.Intn(len(slice))]
	return clientsInfo[ind].pid
}

var detectionMsg string = " "

func connectOthers() {
	for id, cl := range clientsInfo {
		con, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(cl.port))
		if err != nil {
			log.Fatal(err)
		}
		clientsInfo[id].s_conn = con

		go serverConnection(clientsInfo[id].s_conn)

	}
}

func checkVictory() bool {
	return me.pid >= getleaderId()
}
func startElection() {

	if checkVictory() {
		for pid, _ := range clientsInfo {
			leader = true
			detectionMsg = " "
			io.WriteString(serverCons[pid], "c"+strconv.Itoa(me.pid)+"\n")
		}
		fmt.Println("I am the Coordinator ")

	} else {
		for pid, _ := range clientsInfo {
			if pid > me.pid {
				sendCount += 1
				io.WriteString(serverCons[pid], "e"+strconv.Itoa(me.pid)+"\n")
			}
		}
	}
	time.Sleep(time.Duration(5) * time.Microsecond * 100000)
	if okCount == 0 {
		for pid, _ := range clientsInfo {
			leader = true
			detectionMsg = " "
			io.WriteString(serverCons[pid], "c"+strconv.Itoa(me.pid)+"\n")

		}
		fmt.Println("I am the Coordinator ")

	}

}
func handleConnection(conn net.Conn, clientCh chan string, rmCh chan int, allCh chan string) {
	reader := bufio.NewReader(conn)
	id, _ := reader.ReadString('\n')
	str_id := id[:len(id)-1]
	clientCh <- str_id
	pid, _ := strconv.Atoi(str_id)
	serverCons[pid] = conn

	//	clientsInfo[pid].c_conn = conn.(*net.TCPConn)
	defer func() {

		//allCh <- "h"
		if pid == getleaderId() {
			rmCh <- pid
			rand.Seed(time.Now().UnixNano())
			r := rand.Intn(6) + 1

			fmt.Printf("r is : %d\n", r)
			time.Sleep(time.Duration(r) * time.Microsecond * 10000)
			if detectionMsg == " " && leader == true {
				s := "d" + strconv.Itoa(me.pid)

				for _, con := range serverCons {

					io.WriteString(con, s+"\n")

				}
				leader = false
				go startElection()

			}
		} else {
			rmCh <- pid
		}

		//if me.pid == randomDetector() {

		//}

	}()
	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			return
		}
	}
	defer conn.Close()

}

var okCount = 0
var sendCount = 0

func serverConnection(conn net.Conn) {
	io.WriteString(conn, strconv.Itoa(me.pid)+"\n")
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if len(msg) > 0 {
			fmt.Println()
			if string(msg[0]) == "d" {
				fmt.Printf("process with pid " + msg[1:len(msg)-1] + " Detected Abnormality.\n")
				detectionMsg = "d"
			} else if string(msg[0]) == "c" {

				fmt.Printf("New coordinator is : PID " + msg[1:len(msg)-1] + "\n")
				leader = true
				detectionMsg = " "

			} else if string(msg[0]) == "e" {
				fmt.Printf("Election request by PID : " + msg[1:len(msg)-1] + "\n")
				id, _ := strconv.Atoi(msg[1 : len(msg)-1])
				io.WriteString(serverCons[id], "o"+"\n")
				startElection()
			} else if string(msg[0]) == "o" {
				fmt.Printf("PID  : " + msg[1:len(msg)-1] + " STATUS - OK" + "\n")
				okCount += 1
			}

		}
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func main() {
	activate = false
	numberClient := os.Args[1]
	protN := os.Args[2]
	var queued_request [][]int

	// go serverJob()\
	var page0 page // dot notation
	page0.pageNumber = 0
	page0.copySet = append(page0.copySet, 0)
	page0.owner = 0

	var page1 page // dot notation
	page1.pageNumber = 1
	page1.copySet = append(page1.copySet, 1)
	page1.owner = 1

	var page2 page // dot notation
	page2.pageNumber = 2
	page2.copySet = append(page2.copySet, 2)
	page2.owner = 2

	var page3 page // dot notation
	page3.pageNumber = 3
	page3.copySet = append(page3.copySet, 3)
	page3.owner = 3

	var page4 page // dot notation
	page4.pageNumber = 4
	page4.copySet = append(page4.copySet, 4)
	page4.owner = 4

	var page5 page // dot notation
	page5.pageNumber = 5
	page5.copySet = append(page5.copySet, 5)
	page5.owner = 5

	var page6 page // dot notation
	page6.pageNumber = 6
	page6.copySet = append(page6.copySet, 6)
	page6.owner = 6

	var page7 page // dot notation
	page7.pageNumber = 7
	page7.copySet = append(page7.copySet, 7)
	page7.owner = 7

	var page8 page // dot notation
	page8.pageNumber = 8
	page8.copySet = append(page8.copySet, 8)
	page8.owner = 8

	var page9 page // dot notation
	page9.pageNumber = 9
	page9.copySet = append(page9.copySet, 9)
	page9.owner = 9

	var pageList []page
	pageList = append(pageList, page0)
	pageList = append(pageList, page1)
	pageList = append(pageList, page2)
	pageList = append(pageList, page3)
	pageList = append(pageList, page4)
	pageList = append(pageList, page5)
	pageList = append(pageList, page6)
	pageList = append(pageList, page7)
	pageList = append(pageList, page8)
	pageList = append(pageList, page9)

	var addressbook []string
	addressbook = append(addressbook, "localhost:5000")
	addressbook = append(addressbook, "localhost:5001")
	addressbook = append(addressbook, "localhost:5002")
	addressbook = append(addressbook, "localhost:5003")
	addressbook = append(addressbook, "localhost:5004")
	addressbook = append(addressbook, "localhost:5005")
	addressbook = append(addressbook, "localhost:5006")
	addressbook = append(addressbook, "localhost:5007")
	addressbook = append(addressbook, "localhost:5008")
	addressbook = append(addressbook, "localhost:5009")

	var connectionPool = make([]net.Addr, connectionNum)
	pc, err := net.ListenPacket("udp", protN)
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	buffer := make([]byte, 5)
	fmt.Println("Waiting for client...")
	for {

		_, addr, err := pc.ReadFrom(buffer)
		fmt.Println("incoming address", addr)
		rcvMsq := string(buffer)
		fmt.Println("Received: " + rcvMsq)
		s := strings.Split(rcvMsq, ",")
		newid, _ := strconv.Atoi(s[0])
		newmsg, _ := strconv.Atoi(s[1])
		newpage, _ := strconv.Atoi(s[2])

		if newmsg == 0 {
			activate = true
			fmt.Println("The replica server is activated!")

		}
		if err == nil {

			// connectionPool[newid] = addr
			// fmt.Println("connected client : ", connectionPool)
			go add(connectionPool, newid, addr)
			time.Sleep(time.Second * 3)

			if newmsg == 1 { //read
				fmt.Println("message received with code number 1 : read , request from: client ", newid)
				var record []int
				record = append(record, newid)
				record = append(record, newpage)
				record = append(record, newmsg)
				queued_request = append(queued_request, record)
				fmt.Println("queue appended! now queue is :", queued_request)

			} else if newmsg == 2 { //write
				fmt.Println("message received with code number 2 : write , request from: client ", newid)
				var record []int
				record = append(record, newid)
				record = append(record, newpage)
				record = append(record, newmsg)
				queued_request = append(queued_request, record)
				fmt.Println("queue appended! now queue is :", queued_request)

			} else if newmsg == 3 {
				go getleaderId()
				go checkVictory()
			}

			// if condition == false && len(queued_request) != 0 {
			// 	mymsg := strconv.Itoa(queued_request[0]) + "," + "1" // 1= allow
			// 	if _, err := pc.WriteTo([]byte(mymsg), connectionPool[queued_request[0]]); err != nil {
			// 		fmt.Println("error on write: " + err.Error())
			// 	}
			// 	condition = true

			// }

		} else {
			fmt.Println("error: " + err.Error())
		}
		numCli, _ := strconv.Atoi(numberClient)
		if condition == false && len(queued_request) == numCli {
			for len(queued_request) > 1 {
				if queued_request[0][2] == 1 {
					fmt.Println("message received with code number 1 : read , request from: client ", queued_request[0][0])
					pageList[queued_request[0][1]].copySet = append(pageList[queued_request[0][1]].copySet, queued_request[0][0])

					fmt.Println("The page want to read is  :", pageList[queued_request[0][1]])
					mymsg := strconv.Itoa(queued_request[0][0]) + "," + "1" + "," + addressbook[queued_request[0][0]] // 1= read
					fmt.Println("MASTER Read forward to holder:", pageList[queued_request[0][1]].pageNumber)

					if activate == true {
						if _, err := pc.WriteTo([]byte(mymsg), connectionPool[pageList[queued_request[0][1]].pageNumber]); err != nil {
							fmt.Println("error on write: " + err.Error())
						}
					}
					queued_request = queued_request[1:]
				} else if queued_request[0][2] == 2 {
					fmt.Println("message received with code number 2 : write , request from: client ", queued_request[0][0])
					fmt.Println("The page want to write is  :", pageList[queued_request[0][1]])
					if activate == true {
						for i := 0; i < len(pageList[queued_request[0][1]].copySet); i++ {

							mymsg := strconv.Itoa(queued_request[0][0]) + "," + "2" + "," + strconv.Itoa(pageList[queued_request[0][1]].pageNumber)
							if _, err := pc.WriteTo([]byte(mymsg), connectionPool[pageList[queued_request[0][1]].copySet[i]]); err != nil {
								fmt.Println("error on write: " + err.Error())
							} // invalidate
							fmt.Println("receive write confirmation from client", pageList[queued_request[0][1]].pageNumber)

						}
						time.Sleep(time.Second * 2)
						mymsg := strconv.Itoa(queued_request[0][0]) + "," + "4" + "," + addressbook[queued_request[0][0]]
						fmt.Println("MASTER Read forward to holder:", pageList[queued_request[0][1]].pageNumber)
						if _, err := pc.WriteTo([]byte(mymsg), connectionPool[pageList[queued_request[0][1]].pageNumber]); err != nil {
							fmt.Println("error on write: " + err.Error())
						}

					}
					queued_request = queued_request[1:]
					time.Sleep(time.Second * 2)
					fmt.Println("receive write confirmation from client", pageList[queued_request[0][1]].pageNumber)

				}
				pageList[queued_request[0][1]].copySet = nil
				fmt.Println("copyset updated! ")

			}
			if len(queued_request) == 1 {
				if queued_request[0][2] == 1 {
					fmt.Println("message received with code number 1 : read , request from: client ", queued_request[0][0])
					pageList[queued_request[0][1]].copySet = append(pageList[queued_request[0][1]].copySet, queued_request[0][0])
					fmt.Println("The page want to read is  :", pageList[queued_request[0][1]])
					if activate == true {
						mymsg := strconv.Itoa(queued_request[0][0]) + "," + "1" + "," + addressbook[queued_request[0][0]] // 1= read
						fmt.Println("MASTER Read forward to holder:", pageList[queued_request[0][1]].pageNumber)
						if _, err := pc.WriteTo([]byte(mymsg), connectionPool[pageList[queued_request[0][1]].pageNumber]); err != nil {
							fmt.Println("error on write: " + err.Error())
						}
					}
				} else if queued_request[0][2] == 2 {
					fmt.Println("message received with code number 2 : write , request from: client ", queued_request[0][0])
					fmt.Println("The page want to write is  :", pageList[queued_request[0][1]])
					if activate == true {
						for i := 0; i < len(pageList[queued_request[0][1]].copySet); i++ {

							mymsg := strconv.Itoa(queued_request[0][0]) + "," + "2" + "," + strconv.Itoa(pageList[queued_request[0][1]].pageNumber)
							if _, err := pc.WriteTo([]byte(mymsg), connectionPool[pageList[queued_request[0][1]].copySet[i]]); err != nil {
								fmt.Println("error on write: " + err.Error())
							} // invalidate

						}

						time.Sleep(time.Second * 2)

						mymsg := strconv.Itoa(queued_request[0][0]) + "," + "4" + "," + addressbook[queued_request[0][0]]
						fmt.Println("MASTER Read forward to holder:", pageList[queued_request[0][1]].pageNumber)
						if _, err := pc.WriteTo([]byte(mymsg), connectionPool[pageList[queued_request[0][1]].pageNumber]); err != nil {
							fmt.Println("error on write: " + err.Error())

						}
						time.Sleep(time.Second * 2)
						fmt.Println("receive write confirmation from client", pageList[queued_request[0][1]].pageNumber)
					}
					pageList[queued_request[0][1]].copySet = nil
					fmt.Println("copyset updated! ")

				}

			}
		} //else if condition == false && len(queued_request) == numCli {
		// 	fmt.Println(len(queued_request))
		// }
		// 	fmt.Println(len(queued_request))
		// 	for len(queued_request) > 1 {
		// 		fmt.Println("message received with code number 1 : read , request from: client ", queued_request[0][0])
		// 		pageList[queued_request[0][1]].copySet = append(pageList[queued_request[0][1]].copySet, queued_request[0][0])
		// 		fmt.Println("The page want to read is  :", pageList[queued_request[0][1]])
		// 		mymsg := strconv.Itoa(queued_request[0][0]) + "," + "1" + "," + addressbook[queued_request[0][0]] // 1= read
		// 		fmt.Println("MASTER Read forward to holder:", pageList[queued_request[0][1]].pageNumber)
		// 		if _, err := pc.WriteTo([]byte(mymsg), connectionPool[pageList[queued_request[0][1]].pageNumber]); err != nil {
		// 			fmt.Println("error on write: " + err.Error())
		// 		}
		// 		queued_request = queued_request[1:]
		// 	}
		// 	if len(queued_request) == 1 {
		// 		fmt.Println("message received with code number 1 : read , request from: client ", queued_request[0][0])
		// 		pageList[queued_request[0][1]].copySet = append(pageList[queued_request[0][1]].copySet, queued_request[0][0])
		// 		fmt.Println("The page want to read is  :", pageList[queued_request[0][1]])
		// 		mymsg := strconv.Itoa(queued_request[0][0]) + "," + "1" + "," + addressbook[queued_request[0][0]] // 1= read
		// 		fmt.Println("MASTER Read forward to holder:", pageList[queued_request[0][1]].pageNumber)
		// 		if _, err := pc.WriteTo([]byte(mymsg), connectionPool[pageList[queued_request[0][1]].pageNumber]); err != nil {
		// 			fmt.Println("error on write: " + err.Error())
		// 		}

	}
}

func add(connectionPool []net.Addr, newid int, addr net.Addr) {
	connectionPool[newid] = addr
	fmt.Println("connected client : ", connectionPool)
}
