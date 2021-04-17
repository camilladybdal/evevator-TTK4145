package main

import (

	//. "./FSM"
	"./timer"
	//"./elevio"
	//. "./orderDistributor"
	//. "./types"
	"fmt"
	//."./config"
	"time"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func main() {
	test := 10000
	test2 := 100000
	var duration time.Duration
	duration = 1
	newCountdownTime := make(chan time.Duration)
	timedOut := make(chan bool)

	go timer.ResetableTimer(duration, newCountdownTime, timedOut)

	for {
		select {
		case <- timedOut:
			fmt.Println("Timed out")
			//newCountdownTime <- 1
		default:
			if test > 0 {
				newCountdownTime <- 1
				test--
				test2 = 1000000000
			} else {
				//newCountdownTime <- duration
				//fmt.Println("!")
				if test2 < 0 {
					fmt.Println("?")
					test = 100000
				} 
				test2--
			}
		}
	}

	/*
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

	*/
	for {}

}


