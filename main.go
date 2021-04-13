package main

import (
	"./elevio"
	//"./fsm"
	//"./timer"
	. "./orderDistributor"
	. "./types"
)

func main() {

	orderOut := make(chan Order)
	orderIn := make(chan Order)
	getElevatorState := make(chan Elevator)

	elevio.Init("localhost:15657", 4)

	go OrderDistributor(orderOut, orderIn, getElevatorState)

	for {
	}
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