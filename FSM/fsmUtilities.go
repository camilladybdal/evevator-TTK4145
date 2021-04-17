package fsm

import (
	"../elevio"
	. "../types"
	. "../config"
	"fmt"
)

func expidizeOrder(elevator Elevator, OrderUpdate chan<- Order){
	var Expidized_order Order
	Expidized_order.Floor = elevator.CurrentFloor 
	Expidized_order.Status = Done 
	Expidized_order.FromId = ElevatorId
	OrderUpdate <- Expidized_order
}

func getDirection(currentFloor int, destinationFloor int) elevio.MotorDirection {
	if currentFloor-destinationFloor > 0 {
		return elevio.MD_Down
	} else {
		return elevio.MD_Up
	}
}

func checkOrdersPresent(elevator Elevator) bool{
	foundOrder := false
	for i := 0; i < NumberOfFloors; i++ {
		if elevator.UpQueue[i] == 1 || elevator.DownQueue[i] == 1 {
			foundOrder = true
		}
	}
	return foundOrder
}

func queueSearch(QueueDirection elevio.MotorDirection, elevator Elevator) int {
	nextFloor := -1

	if QueueDirection == elevio.MD_Stop{
		QueueDirection = elevio.MD_Up
	}
	
	fmt.Println(" ---- QUEUESEARCH, MY DIRECTION IS: ", QueueDirection);
	fmt.Println(" ---- QUEUESEARCH, MY CUREENT FLOOR IS: ", elevator.CurrentFloor);

	if QueueDirection == elevio.MD_Up {
		for floor := elevator.CurrentFloor; floor < NumberOfFloors; floor++ {
			if elevator.UpQueue[floor] == 1 {
				nextFloor = floor

				return nextFloor
			}
		}
		for floor := NumberOfFloors - 1; floor >= 0; floor-- {
			if elevator.DownQueue[floor] == 1 {
				nextFloor = floor
				return nextFloor
			}
		}
		for floor := 0; floor < elevator.CurrentFloor; floor++ {
			if elevator.UpQueue[floor] == 1 {
				nextFloor = floor
				return nextFloor
			}
		}
	}

	if QueueDirection == elevio.MD_Down{

		for floor := elevator.CurrentFloor; floor >= 0; floor-- {
			if elevator.DownQueue[floor] == 1 {
				nextFloor = floor
				return nextFloor
			}
		}
		for floor := 0; floor < NumberOfFloors; floor++ {
			if elevator.UpQueue[floor] == 1 {
				nextFloor = floor
				return nextFloor
			}
		}
		for floor := NumberOfFloors-1; floor >= 0; floor-- {
			if elevator.DownQueue[floor] == 1 {
				nextFloor = floor
				return nextFloor
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
	for floor := 0; floor < NumberOfFloors; floor++ {
		elevator.UpQueue[floor] = 0
		elevator.DownQueue[floor] = 0
	}
}

