package main

import (

	//"./fsm"
	//"./timer"
	"./elevio"
	. "./orderDistributor"
	. "./types"
	"fmt"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func main() {

	fmt.Println("Hellooooo")
	orderOut := make(chan Order)
	orderIn := make(chan Order)
	getElevatorState := make(chan Elevator)

	elevio.Init("localhost:15657", 4)
	//elevio.Init("10.0.0.5:15658", 4)

	go OrderDistributor(orderOut, orderIn, getElevatorState)
	for {
	}
	//networkTransmit := make(chan Order)
	//networkReceive := make(chan Order)
	/*
		helloTx := make(chan HelloMsg)
		helloRx := make(chan HelloMsg)

		go bcast.Transmitter(config.Port, helloTx)
		go bcast.Receiver(config.Port, helloRx)

		go func() {
			helloMsg := HelloMsg{"Helloooo", 0}
			for {
				helloMsg.Iter++
				helloTx <- helloMsg
				time.Sleep(1 * time.Second)
			}
		}()

		for {
			select {
			case a := <-helloRx:
				fmt.Println("Received: %#v\n", a)
			}
		}*/
	/*
		elevator.UpQueue[0] = 1
		//elevator.UpQueue[2] = 1
		for i := 0; i < NumFloors; i++ {
			elevator.DownQueue[i] = 0
		}
		elevator.DownQueue[3] = 1
		cost := costfnc.Costfunction(elevator, neworder)
		println(cost)
	*/
}

// lage en cannel som sender order, og en som tar i mot
// putt dem i bcast transmit og recieve
//
