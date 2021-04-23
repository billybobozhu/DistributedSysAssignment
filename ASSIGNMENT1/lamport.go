package main

// collaborate with Luo Yifan

import (
	"container/list"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type clock int

const clientNum = 5 //CHANGE NUMBER OF CLIENT HERE

type message struct {
	senderID   int
	senderTime clock
	text       string
}

func listen(serialNum int, chanIN []chan message, chanOUT []chan message) {
	for msg := range chanIN[serialNum] {
		for i := 0; i < cap(chanOUT); i++ {
			if i != serialNum {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500))) // SERVER RANDOM SLEEP FOR A WHILE ZZZ
				chanOUT[i] <- msg
			}
		}
	}
}

func server(portIn []chan message, portOut []chan message) {
	fmt.Println("Server starting")
	for n := 0; n < cap(portIn); n++ {
		go listen(n, portIn, portOut)
	}

}

func client(serialNum int, portIn chan message, portOut chan message, messageList *list.List) {
	fmt.Printf("Client %d starting\n", serialNum)
	var localClock clock = 0
	var lock = new(sync.RWMutex)
	go Lamport(serialNum, portIn, messageList, lock, &localClock)
	for i := 0; i < 10; i++ { //GENERATE 10 MESSAGES AND SEND TO SERVER
		lock.Lock()
		localClock++
		msg := message{
			serialNum,
			localClock,
			fmt.Sprintf("%d", rand.Intn(100))}
		lock.Unlock()

		portOut <- msg //AND SEND TO SERVER
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
}
func Lamport(serialNum int, chanIN chan message, messageList *list.List, lock *sync.RWMutex, clk *clock) {
	for msg := range chanIN {
		lock.Lock()

		if msg.senderTime > *clk {
			*clk = msg.senderTime + 1 //LAMPORT INCREMENT CLK
		} else {
			*clk++
		}

		messageList.PushBack(msg) //ADD MESSAGE TO THE QUEUE
		fmt.Printf("client ID  %d and senderclock %d message received: %+v\n", serialNum, *clk, msg)
		lock.Unlock()
	}
}

func main() {
	fmt.Println("start")

	var chStoC [clientNum]chan message
	var chCtoS [clientNum]chan message
	var messageLists [clientNum]*list.List //INIT CHANNEL

	var waitgroup sync.WaitGroup

	for i := 0; i < clientNum; i++ {
		chStoC[i] = make(chan message)
		chCtoS[i] = make(chan message)
		messageLists[i] = list.New() //INIT MESSAGE QUEUE
	}

	go server(chCtoS[:], chStoC[:])
	fmt.Println("server start")

	for c := 0; c < clientNum; c++ {
		waitgroup.Add(1)
		go func(c int) {
			client(c, chStoC[c], chCtoS[c], messageLists[c])
			waitgroup.Done()
		}(c)
	}
	waitgroup.Wait()

	time.Sleep(time.Duration(1000) * time.Millisecond)
	for c := 0; c < clientNum; c++ {
		fmt.Printf("\nClient %d:\n", c)
		for log := messageLists[c].Front(); log != nil; log = log.Next() { //read according to sender time
			fmt.Printf("%+v\n", log.Value)
		}
	}

	var input string
	fmt.Scanln(&input)
}
