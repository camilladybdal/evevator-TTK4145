package main

import (
	"fmt"
	"time"

	. "./config"
	"./elevio"
	. "./fsm"
	. "./orderDistributor"
	. "./types"
)

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

	elevio.Init(ElevatorAddress, NumberOfFloors)
	InitFSM(NumberOfFloors)

	go OrderDistributor(fsmChannels.NewOrder, orderUpdate, getElevatorState)
	go RunElevator(fsmChannels, orderUpdate, getElevatorState)

	for {
		time.Sleep(3 * time.Second)
		fmt.Println("MAIN RUNNING")
	}

}
