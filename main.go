package main

import (

	. "./fsm"
	//"./timer"
	"./elevio"
	. "./orderDistributor"
	. "./types"
	"fmt"
	."./config"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func main() {

	fmt.Println("LETS GO")
	
	// FSM channels
	var fsmChannels FsmChannels
	fsmChannels.FloorReached = make(chan int)
	fsmChannels.NewOrder = make(chan Order)
	fsmChannels.Obstruction = make(chan bool)
	fsmChannels.ElevatorState = make(chan Elevator)
	fsmChannels.DoorTimedOut = make(chan bool)
	fsmChannels.Immobile = make(chan int)
	fsmChannels.StopImmobileTimer = make(chan bool)

	//shared channels
	orderUpdate := make(chan Order)
	getElevatorState := make(chan Elevator)


	elevio.Init("localhost:15657", NumberOfFloors)
	InitFSM(NumberOfFloors)
	//elevio.Init("10.0.0.5:15658", 4)

	go OrderDistributor(fsmChannels.NewOrder, orderUpdate, getElevatorState)
	go RunElevator(fsmChannels, orderUpdate, getElevatorState)


	for {}

}


