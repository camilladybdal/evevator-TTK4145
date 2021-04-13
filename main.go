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

}
