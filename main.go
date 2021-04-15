package main

import (
	"./elevio"
	"./fsm"
	//"./timer"
	. "./types"
	//"fmt"
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
	order3.Floor = 1
	order3.DirectionUp = true
	order3.CabOrder = false

	var order4 Order
	order4.Floor = 3
	order4.DirectionDown = true
	order4.CabOrder = false

	var order5 Order
	order5.Floor = 0
	order5.DirectionUp = false
	order5.CabOrder = true

	var order6 Order
	order6.Floor = 2
	order6.DirectionUp = true
	order6.CabOrder = false

	channels.NewOrder <- order1
	channels.NewOrder <- order2
	time.Sleep(4*time.Second)
	channels.NewOrder <- order3
	time.Sleep(20*time.Second)
	channels.NewOrder <- order4
	channels.NewOrder <- order5
	channels.NewOrder <- order6

	select{}

}
