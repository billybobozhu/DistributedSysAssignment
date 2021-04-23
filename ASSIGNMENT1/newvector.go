package main

import (
	"container/list"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const numProcesses = 3 //DEFINE NO. OF CLIENTS

type message struct {
	senderID  int    // ORIGIN OF MESSAGE
	vectorCLK []int  //VECTOR CLOCK  [0 0 0 0]
	random    string //RANDOM MESSAGE
}

func listen(serialNum int, chIn []chan message, chOut []chan message, localvector []int) {
	//SERVER WILL RECEIVE ALL THE MESSAGES FROM CLIENTS
	var lock = new(sync.RWMutex)
	for msg := range chIn[serialNum] { // MSG INCOMING FROM A CLIENT

		lock.Lock()
		var random = msg.random //TAKE OUT RANDOM MSG
		for i := 0; i < numProcesses+1; i++ {
			if msg.vectorCLK[i] > localvector[i] {
				localvector[i] = msg.vectorCLK[i] //UPDATE THE VECTOR ACCORDING TO SERVER MAINTAINED CLK

			}

		}
		localvector[len(localvector)-1]++ //SERVER EVENT ++

		fmt.Printf("SERVER MSG: message from client %d received: %+v\n", serialNum, msg)

		for i := 0; i < cap(chOut); i++ {

			if i != serialNum { //SEND TO OTHER CLIENTS

				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500))) //RANDOM DELAY

				localvector[len(localvector)-1]++ //SERVER EVENT ++
				newmsg := message{
					serialNum,
					localvector, //USE UPDATED LOCAL VECTOR
					random}      //PUT IN ORIGNAL RANDOM

				chOut[i] <- newmsg //SEND OUT TO OTHER CLIENTS
				fmt.Printf("SERVER MSG: server sent to client%d with %+v\n ", i, newmsg)

			}
		}
		lock.Unlock()
	}
}

func server(portIn []chan message, portOut []chan message) {
	fmt.Println("Server starting")
	var localvector = make([]int, numProcesses+1) //INIT
	for n := 0; n < cap(portIn); n++ {
		go listen(n, portIn, portOut, localvector)
	}

}

func client(serialNum int, portIn chan message, portOut chan message, messageList *list.List) {
	var clockTime = time.Duration(500+rand.Intn(1000)) * time.Millisecond //RANDOM DELAY

	fmt.Printf("Client %d starting\n", serialNum)
	var localvector = make([]int, numProcesses+1)
	var lock = new(sync.RWMutex)
	go vector(serialNum, portIn, messageList, lock, localvector)

	for i := 0; i < 5; i++ { //GO FOR 5 GENERATION OF EVENTS

		lock.Lock()
		var random = fmt.Sprintf("%d", rand.Intn(100))
		localvector[serialNum]++ //INCREMENT VECTOR (FOR SENDING MESSAGE)
		msg := message{
			serialNum,
			localvector,
			random}
		lock.Unlock()

		portOut <- msg
		fmt.Printf("client %d send message to server: %+v\n", serialNum, msg)
		time.Sleep(clockTime)
	}
}
func vector(serialNum int, chIn chan message, messageList *list.List, lock *sync.RWMutex, localvector []int) {
	for msg := range chIn {
		lock.Lock()
		messageList.PushBack(msg)
		lock.Unlock()
		lock.Lock()

		for i := 0; i < numProcesses+1; i++ {
			if msg.vectorCLK[i] > localvector[i] {
				localvector[i] = msg.vectorCLK[i] //UPDATE LOCAL VECTOR

			}
			if msg.vectorCLK[i] < localvector[i] {
				fmt.Println("error!!!") //If the local clock is after (in line with the vector comparison) the message clock, then a causality violation is flagged.

			}

		}
		localvector[serialNum]++ //INCREMENT VECTOR (FOR RECEIVING MESSAGE)

		fmt.Printf("serialNum %d message received: %+v\n", serialNum, msg)
		lock.Unlock()
	}
}

// func showLocalVector(serialnumber int, vector []int) {
// 	fmt.Printf("Process %d has vectorclock %+v \n", serialnumber, vector)
// }

func main() {
	fmt.Println("start")

	var chStoC [numProcesses]chan message
	var chCtoS [numProcesses]chan message
	var messageLists [numProcesses]*list.List //INIT

	var waitgroup sync.WaitGroup

	for i := 0; i < numProcesses; i++ {
		chStoC[i] = make(chan message)
		chCtoS[i] = make(chan message)
		messageLists[i] = list.New()
	}

	go server(chCtoS[:], chStoC[:]) //START SERVER

	for c := 0; c < numProcesses; c++ {
		waitgroup.Add(1)
		go func(c int) {
			client(c, chStoC[c], chCtoS[c], messageLists[c]) //START CLIENT
			waitgroup.Done()
		}(c)
	}
	waitgroup.Wait()

	// for c := 0; c < numProcesses; c++ {
	// 	fmt.Printf("\nClient %d:\n", c)
	// 	for log := messageLists[c].Front(); log != nil; log = log.Next() {
	// 		fmt.Printf("%+v\n", log.Value)
	// 	}
	// }

	// var input string
	// fmt.Scanln(&input)
}
