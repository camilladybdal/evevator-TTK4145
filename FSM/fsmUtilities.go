package fsm

import (
	"../elevio"
	. "../types"
)

func getDirection(currentFloor int, destinationFloor int) elevio.MotorDirection {
	if currentFloor-destinationFloor > 0 {
		return elevio.MD_Down
	} else {
		return elevio.MD_Up
	}
}

func checkOrdersPresent(elevator Elevator) bool{
	foundOrder := false
	for i := 0; i < NumFloors; i++ {
		if elevator.UpQueue[i] == 1 || elevator.DownQueue[i] == 1 {
			foundOrder = true
		}
	}
	return foundOrder
}

func queueSearch(QueueDirection elevio.MotorDirection, elevator Elevator) int {
	nextFloor := 0

	//first times
	if QueueDirection == elevio.MD_Stop{
		QueueDirection = elevio.MD_Up
	}


	if QueueDirection == elevio.MD_Up {
		for floor := elevator.CurrentFloor; floor < NumFloors; floor++ {
			if elevator.UpQueue[floor] == 1 {
				nextFloor = floor
				break
			}
		}
		for floor := NumFloors - 1; floor >= 0; floor-- {
			if elevator.DownQueue[floor] == 1 {
				nextFloor = floor
				break
			}
		}
		for floor := 0; floor < elevator.CurrentFloor; floor++ {
			if elevator.UpQueue[floor] == 1 {
				nextFloor = floor
				break
			}
		}
	}

	if QueueDirection == elevio.MD_Down{
		for floor := elevator.CurrentFloor; floor >= 0; floor-- {
			if elevator.DownQueue[floor] == 1 {
				nextFloor = floor
				break
			}
		}
		for floor := 0; floor < NumFloors; floor++ {
			if elevator.UpQueue[floor] == 1 {
				nextFloor = floor
				break
			}
		}
		for floor := elevator.CurrentFloor; floor >= 0; floor-- {
			if elevator.DownQueue[floor] == 1 {
				nextFloor = floor
				break
			}
		}
	}
	return nextFloor
}

func removeFromQueue(elevator *Elevator){
		elevator.UpQueue[elevator.CurrentFloor] = 0
		elevator.DownQueue[elevator.CurrentFloor] = 0
}

func emptyQueue(elevator *Elevator){
	for floor := 0; floor < NumFloors; floor++ {
		elevator.UpQueue[elevator.CurrentFloor] = 0
		elevator.DownQueue[elevator.CurrentFloor] = 0
	}
}