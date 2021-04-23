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

var holdingReadonly []int
var holdingRW []int
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

// func remove(s []int, i int) []int {
// 	if len(s) == 0 && i == 0 {
// 		s = nil
// 		return s
// 	}
// 	s[i] = s[len(s)-1]
// 	// We do not need to put s[i] at the end, as it will be discarded anyway
// 	return s[:len(s)-1]
// }

func server(conn *net.UDPConn, listenport string, service2 string, id int) {
	for {
		buffer := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Master Server Address: ", addr)
		fmt.Println("Received from server : ", string(buffer[:n]))
		k := string(buffer[:n])

		s := strings.Split(k, ",")

		newid, _ := strconv.Atoi(s[0])
		reply, _ := strconv.Atoi(s[1])
		fmt.Println(s[2])

		if reply == 1 { //

			// time.Sleep(time.Second * 10)
			fmt.Printf("receive read request from :%d\n", newid)
			time.Sleep(time.Second * 1)
			fmt.Printf("sending page to the target :%d\n", newid)
			now := time.Now()
			nsec := now.UnixNano()
			fmt.Println(nsec)

			strId := strconv.Itoa(newid)

			// var pageName string
			ids := strconv.Itoa(id)
			strMessage := fmt.Sprintf("%s%s", "page", ids)
			//E concatenar todas
			mymsg := strId + "," + ids + "," + strMessage + "," + "read"

			// buf := []byte(mymsg)
			// replyaddr, _, _ := net.ParseCIDR(s[2])
			// fmt.Println("Requess:=strconvt macheine addr is :", replyaddr)
			// addr := &net.IPAddr{replyaddr, ""}
			// service3 := "localhost:" + listenport
			// service4 := "localhost:" + strconv.Itoa(5998)
			// fmt.Println(service3)
			RemoteAddr, err := net.ResolveUDPAddr("udp", s[2])
			LocalAddr, err := net.ResolveUDPAddr("udp", service2)

			conn2, err := net.DialUDP("udp", LocalAddr, RemoteAddr)

			if err != nil {
				log.Fatal(err)
			}

			// log.Printf("Established connection to %s \n", s[2])
			// log.Printf("Remote UDP address : %s \n", conn2.RemoteAddr().String())
			// log.Printf("Local UDP client address : %s \n", conn2.LocalAddr().String())

			defer conn2.Close()

			buf := []byte(mymsg)

			_, err = conn2.Write(buf)
			if err != nil {
				fmt.Println(mymsg, err)
			}

		} else if reply == 3 {
			fmt.Printf("receive page  from !\n", newid)
			time.Sleep(time.Second * 1)
			fmt.Println(s[2])

		} else if reply == 2 {
			fmt.Println("receive messgae  from master to invalidate page")

			time.Sleep(time.Second * 1)
			deletepage, _ := strconv.Atoi(s[2])

			for i := 0; i < len(holdingReadonly); i++ {

				if holdingReadonly[i] == deletepage {
					holdingReadonly = nil
				}
			}

			fmt.Println("holding list update:", holdingReadonly)

		} else if reply == 4 {
			// time.Sleep(time.Second * 10)
			fmt.Printf("receive read request from :%d\n", newid)
			time.Sleep(time.Second * 1)
			fmt.Printf("sending page to the target :%d\n", newid)
			now := time.Now()
			nsec := now.UnixNano()
			fmt.Println(nsec)

			strId := strconv.Itoa(newid)

			// var pageName string
			ids := strconv.Itoa(id)
			strMessage := fmt.Sprintf("%s%s", "page", ids)
			//E concatenar todas
			mymsg := strId + "," + ids + "," + strMessage + "," + "write"

			// buf := []byte(mymsg)
			// replyaddr, _, _ := net.ParseCIDR(s[2])
			// fmt.Println("Requess:=strconvt macheine addr is :", replyaddr)
			// addr := &net.IPAddr{replyaddr, ""}
			// service3 := "localhost:" + listenport
			// service4 := "localhost:" + strconv.Itoa(5998)
			// fmt.Println(service3)
			RemoteAddr, err := net.ResolveUDPAddr("udp", s[2])
			LocalAddr, err := net.ResolveUDPAddr("udp", service2)

			conn2, err := net.DialUDP("udp", LocalAddr, RemoteAddr)

			if err != nil {
				log.Fatal(err)
			}

			// log.Printf("Established connection to %s \n", s[2])
			// log.Printf("Remote UDP address : %s \n", conn2.RemoteAddr().String())
			// log.Printf("Local UDP client address : %s \n", conn2.LocalAddr().String())

			defer conn2.Close()

			buf := []byte(mymsg)

			_, err = conn2.Write(buf)
			if err != nil {
				fmt.Println(mymsg, err)
			}

		}
	}

}

func main() {

	inCritical = false
	waIting = false
	hostName := "localhost"
	portNum := "6000"
	id, _ = strconv.Atoi(os.Args[1])

	holdingReadonly = append(holdingReadonly, id)
	myPort = os.Args[2]
	pageId := os.Args[3]
	listenport := os.Args[4]
	readorwrite := os.Args[5]

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

	// write a message to server
	for {

		go server(conn, listenport, service2, id)
		go listen(listenport, readorwrite)

		if readorwrite == "read" {
			time.Sleep(time.Second * 3)

			fmt.Printf("Requesting read with ID = %d \n", id)
			now := time.Now()
			nsec := now.UnixNano()
			fmt.Println(nsec)
			// text_simples := "CS here"

			strId := strconv.Itoa(id)
			strMessage := strconv.Itoa(1)
			// strPage := strconv.Itoa(method)
			//E concatenar todas
			mymsg := strId + "," + strMessage + "," + pageId

			buf := []byte(mymsg)

			_, err = conn.Write(buf)
			if err != nil {
				fmt.Println(mymsg, err)
			}
			time.Sleep(time.Second * 10000)
		} else if readorwrite == "write" {
			time.Sleep(time.Second * 3)
			fmt.Printf("Requesting write with ID = %d \n", id)

			strId := strconv.Itoa(id)
			strMessage := strconv.Itoa(2)
			// strPage := strconv.Itoa(method)
			//E concatenar todas
			mymsg := strId + "," + strMessage + "," + pageId
			fmt.Println("sending write request to master", mymsg)
			buf := []byte(mymsg)

			_, err = conn.Write(buf)
			if err != nil {
				fmt.Println(mymsg, err)
			}
			time.Sleep(time.Second * 10000)

		}
	}

}

func readin() {
	fmt.Scanln()
}

func listen(port string, method string) {
	fmt.Println("listening at port:", port)
	pc, err := net.ListenPacket("udp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	buffer := make([]byte, 12)
	fmt.Println("Waiting for client...")
	for {

		_, addr, _ := pc.ReadFrom(buffer)
		fmt.Println("incoming address", addr)
		rcvMsq := string(buffer)
		fmt.Println("Received: " + rcvMsq)
		s := strings.Split(rcvMsq, ",")

		fmt.Printf("Page received from client %s content is %s", s[0], s[2])
		if method == "write" {
			fmt.Println(s)
			holdpage, _ := strconv.Atoi(s[1])
			holdingRW = append(holdingRW, holdpage)
			fmt.Println("client now holding R&W page :", holdingRW)

		} else {
			fmt.Println(s)
			holdpage, _ := strconv.Atoi(s[1])
			holdingReadonly = append(holdingReadonly, holdpage)
			fmt.Println("client now holding read only page :", holdingReadonly)

		}

	}
}
