package main

import (
	"./elevio"
	"./fsm"
	//"./timer"
	. "./types"
	"fmt"
	"time"
)



func main() {

	elevio.Init("localhost:15657", NumFloors)
	fsm.InitFSM(2)


	time.Sleep(5*time.Second)
	
	var channels FsmChannels
	channels.NewOrder = make(chan Order)
	channels.FloorReached = make(chan int)
	channels.Obstruction = make(chan bool)
	channels.ElevatorState  = make(chan Elevator)
	channels.DoorTimedOut = make(chan bool)
	channels.Immobile = make(chan int)
	channels.StopImmobileTimer = make(chan bool)

	OrderUpdate := make(chan Order)
	elevInfo := make(chan Elevator)

	go fsm.RunElevator(channels, OrderUpdate, elevInfo)


	//lage test-ordre
	var order1 Order
	order1.Floor = 3
	order1.DirectionDown = true
	order1.CabOrder = false

	var order2 Order
	order2.Floor = 2
	order2.DirectionUp = true
	order2.CabOrder = false

	var order3 Order
	order3.Floor = 0
	order3.DirectionUp = true
	order3.CabOrder = false


	fmt.Println("sending")
	channels.NewOrder <- order1
	fmt.Println("sent")

	fmt.Println("sending")
	channels.NewOrder <- order2
	fmt.Println("sent")

	fmt.Println("sending")
	channels.NewOrder <- order3
	fmt.Println("sent")

	select{}

}
