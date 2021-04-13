package main

import (
	"./costfnc"
	. "./types"
)

func main() {
	
	var neworder Order
	neworder.Floor = 1
	neworder.DirectionUp = true
	var elevator Elevator
	elevator.CurrentFloor = 2
	elevator.Direction = 0
	for i := 0; i < NumFloors; i++ {
		elevator.UpQueue[i] = 0
	}
	elevator.UpQueue[0] = 1
	//elevator.UpQueue[2] = 1
	for i := 0; i < NumFloors; i++ {
		elevator.DownQueue[i] = 0
	}
	elevator.DownQueue[3] = 1
	cost := costfnc.Costfunction(elevator, neworder)
	println(cost)
}
