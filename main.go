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

	// OrderDistributor channels
	var orderDistributorChannels OrderDistributorChannels
	orderDistributorChannels.NewButtonEvent = make(chan elevio.ButtonEvent)
	orderDistributorChannels.OrderUpdate = make(chan Order)
	orderDistributorChannels.OrderTransmitter = make(chan Order)
	orderDistributorChannels.OrderReciever = make(chan Order)

	//shared channels
	
	//getElevatorState := make(chan Elevator)

	elevio.Init(ElevatorAddress, NumberOfFloors)
	go elevio.PollFloorSensor(fsmChannels.FloorReached)
	go elevio.PollObstructionSwitch(fsmChannels.Obstruction)
	go elevio.PollButtons(orderDistributorChannels.NewButtonEvent)

	go bcast.Transmitter(Port, orderDistributorChannels.OrderTransmitter)
	go bcast.Receiver(Port, orderDistributorChannels.OrderReciever)
	
	InitFSM(NumberOfFloors)

	//go OrderDistributor(fsmChannels.NewOrder, orderUpdate, fsmChannels.ElevatorState, newButtonEvent, networkTransmit, networkRecieve)
	go OrderDistributor(orderDistributorChannels, fsmChannels.NewOrder, fsmChannels.ElevatorState)
	
	go RunElevator(fsmChannels, orderDistributorChannels.OrderUpdate)

	for {
		time.Sleep(3 * time.Second)
		fmt.Println("MAIN RUNNING")
	}

}
