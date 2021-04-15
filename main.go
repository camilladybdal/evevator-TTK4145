package main

import (

	. "./FSM"
	//"./timer"
	"./elevio"
	. "./orderDistributor"
	. "./types"
	"fmt"
	. "./config"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func main() {

	fmt.Println("Hellooooo")
	// orderOut := make(chan Order)
	// orderIn := make(chan Order)
	

	// FSM channels
	var fsmChannels FsmChannels
	fsmChannels.FloorReached = make(chan int)
	fsmChannels.NewOrder = make(chan Order)
	fsmChannels.Obstruction = make(chan bool)
	fsmChannels.ElevatorState = make(chan Elevator)
	fsmChannels.DoorTimedOut = make(chan bool)
	fsmChannels.Immobile = make(chan int)
	fsmChannels.StopImmobileTimer = make(chan bool)

	orderUpdate := make(chan Order)
	getElevatorState := make(chan Elevator)


	elevio.Init("localhost:15657", NumberOfFloors)
	InitFSM(NumberOfFloors)
	//elevio.Init("10.0.0.5:15658", 4)

	go OrderDistributor(fsmChannels.NewOrder, orderUpdate, getElevatorState)
	go RunElevator(fsmChannels, orderUpdate, getElevatorState)
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
