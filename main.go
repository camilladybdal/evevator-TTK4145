package main

import (
	"fmt"
	"time"
	"./network/bcast"
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

	networkTransmit := make(chan Order)
	networkRecieve := make(chan Order)

	//shared channels
	
	//getElevatorState := make(chan Elevator)

	elevio.Init(ElevatorAddress, NumberOfFloors)
	go elevio.PollFloorSensor(fsmChannels.FloorReached)
	go elevio.PollObstructionSwitch(fsmChannels.Obstruction)
	go elevio.PollButtons(newButtonEvent)

	go bcast.Transmitter(Port, networkTransmit)
	go bcast.Receiver(Port, networkRecieve)
	


	InitFSM(NumberOfFloors)

	go OrderDistributor(fsmChannels.NewOrder, orderUpdate, fsmChannels.ElevatorState, newButtonEvent, networkTransmit, networkRecieve)
	
	
	go RunElevator(fsmChannels, orderUpdate)

	for {
		time.Sleep(3 * time.Second)
		fmt.Println("MAIN RUNNING")
	}

}
