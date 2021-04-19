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

	newButtonEvent := make(chan elevio.ButtonEvent)
	orderUpdate := make(chan Order)

	//shared channels
	
	//getElevatorState := make(chan Elevator)

	elevio.Init(ElevatorAddress, NumberOfFloors)
	go elevio.PollFloorSensor(fsmChannels.FloorReached)
	go elevio.PollObstructionSwitch(fsmChannels.Obstruction)
	go elevio.PollButtons(newButtonEvent)
	


	InitFSM(NumberOfFloors)

	go OrderDistributor(fsmChannels.NewOrder, orderUpdate, fsmChannels.ElevatorState, newButtonEvent)
	
	
	go RunElevator(fsmChannels, orderUpdate)

	for {
		time.Sleep(3 * time.Second)
		fmt.Println("MAIN RUNNING")
	}

}
