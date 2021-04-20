package main

import (
	"fmt"

	. "./config"
	"./elevio"
	. "./fsm"
	"./network/bcast"
	. "./orderDistributor"
	. "./types"
)

func main() {

	fmt.Println("LETS GO")

	var fsmChannels FsmChannels
	fsmChannels.FloorReached = make(chan int)
	fsmChannels.NewOrder = make(chan Order)
	fsmChannels.Obstruction = make(chan bool)
	fsmChannels.ElevatorState = make(chan Elevator)
	fsmChannels.DoorTimedOut = make(chan bool)
	fsmChannels.Immobile = make(chan int)
	fsmChannels.StopImmobileTimer = make(chan bool)

	var orderDistributorChannels OrderDistributorChannels
	orderDistributorChannels.NewButtonEvent = make(chan elevio.ButtonEvent)
	orderDistributorChannels.OrderUpdate = make(chan Order)
	orderDistributorChannels.OrderTransmitter = make(chan Order)
	orderDistributorChannels.OrderReciever = make(chan Order)

	elevio.Init(ELEVATOR_ADDRESS, NUMBER_OF_FLOORS)
	go elevio.PollFloorSensor(fsmChannels.FloorReached)
	go elevio.PollObstructionSwitch(fsmChannels.Obstruction)
	go elevio.PollButtons(orderDistributorChannels.NewButtonEvent)

	go bcast.Transmitter(PORT, orderDistributorChannels.OrderTransmitter)
	go bcast.Receiver(PORT, orderDistributorChannels.OrderReciever)

	InitFSM(NUMBER_OF_FLOORS)

	go OrderDistributor(orderDistributorChannels, fsmChannels.NewOrder, fsmChannels.ElevatorState)
	go RunElevator(fsmChannels, orderDistributorChannels.OrderUpdate)

	select {}
}
