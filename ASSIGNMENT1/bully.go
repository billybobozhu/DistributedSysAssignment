package main

import (
	"fmt"
	"sync"
	"time"
)

const numProcesses = 5

type messageType int32

const (
	VICTORY  messageType = 0
	ELECTION messageType = 1
	ALIVE    messageType = 2
)

type aliveMessage struct {
	processID   int
	coordinator bool
}
type message struct {
	processID int
	msgType   messageType
}

var lock sync.Mutex
var coordinator int
var clientnumber int
var coordinatorCondition bool
var receiveVictory []string

func client(id int, chanIn []chan aliveMessage, chanOut []chan message) {
	fmt.Printf("This is Client %d , I am ready \n", id)
	fmt.Printf("client %d knows the coordinator is %d \n", id, coordinator)

	if id == coordinator {
		fmt.Printf("The coordinator is down\n") // here simulate the coordinator is down
		amt := time.Duration(100000000)         //
		time.Sleep(time.Second * amt)           //
		coordinatorCondition = false            //

		for msg := range chanIn[coordinator] {
			fmt.Println("coordinator is alive") //coordinator behavior
			if msg.coordinator == false {
				coordinatorCondition = true

			}
		}

	}
	// if id != 0 { //use this when simulate worst condition
	// 	amt3 := time.Duration(3)
	// 	time.Sleep(time.Second * amt3)
	// }
	// if id != coordinator-1 { //use this when simulate best condition
	// 	amt3 := time.Duration(5)
	// 	time.Sleep(time.Second * amt3)

	// }
	go checkCoordinator(id, chanIn)

	amt1 := time.Duration(5)
	time.Sleep(time.Second * amt1)

	go receive(id, chanIn)
	if coordinatorCondition == false { //continue to election if coordinator is down
		amt2 := time.Duration(3)
		time.Sleep(time.Second * amt2)

		go checkLeader(id, chanOut)

		amt := time.Duration(15)
		time.Sleep(time.Second * amt)
		fmt.Printf("The new coordinator is %d\n ", coordinator)
	} else {
		fmt.Println("coordinator is OK no need to do election")
	}
}

func receive(id int, changive []chan aliveMessage) {

	if coordinatorCondition == true {
		fmt.Println("Coordinator is alive")
	} else {
		fmt.Println("Coordinator is dead and starts election")

	}

}
func checkCoordinator(id int, chanIn []chan aliveMessage) {
	fmt.Printf("Process %d start to check coordinator condition\n", id)
	coordinatorCondition = false
	msg := aliveMessage{
		id,
		false}
	chanIn[coordinator] <- msg

}
func checkLeader(id int, chanOut []chan message) {
	// if id == 3 { // here simulate there is a process suddenly fail during the election, this value can be different.
	// 	fmt.Printf("The process with id %d is dead\n", id)
	// 	amt := time.Duration(100000000)
	// 	time.Sleep(time.Millisecond * amt)

	// }
	go handleMsg(id, chanOut)
	amt := time.Duration(1)
	time.Sleep(time.Second * amt)

	go election(id, chanOut)

}
func election(id int, chanOut []chan message) {
	fmt.Printf("process %d find coordinator %d is down, go for election \n", id, coordinator)

	if id == coordinator-1 && coordinatorCondition == false {
		fmt.Printf("Process %d has the largest id, I am the new coordinator \n", id)
		coordinator = id
		coordinatorCondition = true
		victorymsg := message{
			id,
			VICTORY}
		for i := 0; i < numProcesses-1; i++ {

			if i != id {
				fmt.Printf("Process %d send VICTORY message to process %d\n", id, i)
				chanOut[i] <- victorymsg

			}

		}

	} else {

		electionmsg := message{
			id,
			ELECTION}
		for i := id + 1; i < numProcesses-2; i++ {
			fmt.Printf("Process %d send ELECTION message to process %d\n", id, i)
			chanOut[i] <- electionmsg
		}

		amt := time.Duration(10)
		time.Sleep(time.Second * amt)

		if receiveVictory[id] == "no response" && coordinatorCondition == false {
			fmt.Printf("process %d successfully elected as coordinator\n", id)
			coordinator = id
			victorymsg := message{
				id,
				VICTORY}
			for i := 0; i < numProcesses-1; i++ {
				if i != id {
					fmt.Printf("Process %d send VICTORY message to process %d\n", id, i)
					chanOut[i] <- victorymsg
				}
			}
			coordinatorCondition = true
		}
	}
}

func handleMsg(id int, chanOut []chan message) {
	fmt.Printf("Process %d handleMsg starting\n", id)
	for msg := range chanOut[id] {
		if msg.msgType == VICTORY {
			fmt.Printf("Process %d received VICTORY from %d and understand it is the new coordinator \n", id, msg.processID)
			coordinator = msg.processID
			lock.Lock()
			receiveVictory[id] = "received"
			lock.Unlock()
		} else if msg.msgType == ELECTION {
			if msg.processID < id {
				fmt.Printf("process %d received ELECTION from %d and it is smaller than current process \n", id, msg.processID)

				for i := id + 1; i < numProcesses; i++ {
					aliveMsg := message{
						id,
						ALIVE}
					chanOut[msg.processID] <- aliveMsg
				}
			} else {
				fmt.Printf("process %d received ELECTION from %d and it is larger \n", id, msg.processID)

			}
		} else if msg.msgType == ALIVE {
			if msg.processID > id {
				fmt.Printf("process %d received Alive from %d and it is larger, process start to wait for victory \n", id, msg.processID)
				lock.Lock()
				receiveVictory[id] = "null"
				lock.Unlock()

			}

		}
	}
}

func main() {
	fmt.Println("start")
	coordinator = numProcesses - 1
	for i := 0; i < numProcesses; i++ {
		receiveVictory = append(receiveVictory, "no response")

	}

	var chStoC [numProcesses]chan aliveMessage
	var chCtoS [numProcesses]chan message

	var waitgroup sync.WaitGroup

	for i := 0; i < numProcesses; i++ {
		chStoC[i] = make(chan aliveMessage, 10)
		chCtoS[i] = make(chan message, 10)
	}

	for c := 0; c < numProcesses; c++ {
		waitgroup.Add(1)
		go func(c int) {
			client(c, chStoC[:], chCtoS[:]) //START CLIENT
			waitgroup.Done()
		}(c)
	}
	waitgroup.Wait()

	var input string
	fmt.Scanln(&input)
}
