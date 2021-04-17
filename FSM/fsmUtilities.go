package fsm

import (
	"os"

	"strconv"

	. "../config"
	"../elevio"
	. "../types"
)

func expidizeOrder(elevator Elevator, OrderUpdate chan<- Order) {
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

func checkOrdersPresent(elevator Elevator) bool {
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

	//first times
	if QueueDirection == elevio.MD_Stop {
		QueueDirection = elevio.MD_Up
	}

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

	if QueueDirection == elevio.MD_Down {

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
		for floor := NumberOfFloors - 1; floor >= 0; floor-- {
			if elevator.DownQueue[floor] == 1 {
				nextFloor = floor
				return nextFloor
			}
		}
	}
	return nextFloor
}

func removeFromQueue(elevator *Elevator) {
	elevator.UpQueue[elevator.CurrentFloor] = 0
	elevator.DownQueue[elevator.CurrentFloor] = 0
}

func emptyQueue(elevator *Elevator) {
	for floor := 0; floor < NumberOfFloors; floor++ {
		elevator.UpQueue[floor] = 0
		elevator.DownQueue[floor] = 0
	}
}

////////////FILE/////////////
func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func writeToBackUpFile(filename string, elevatorID int, elevator Elevator) {
	id := strconv.Itoa(elevatorID)
	file, err := os.Create(filename + id)
	checkError(err)
	var write string
	for i := 0; i < NumberOfFloors; i++ {
		if elevator.UpQueue[i] == 1 && elevator.DownQueue[i] == 1 {
			write = write + "1"
		} else {
			write = write + "0"
		}
	}
	data := []byte(write)
	file.Write(data)
}

func readFromBackupFile(filename string, elevatorid int, elevator *Elevator) {
	id := strconv.Itoa(elevatorid)
	file, _ := os.Open(filename + id)
	data := make([]byte, NumberOfFloors)
	file.Read(data)

	for i := 0; i < NumberOfFloors; i++ {
		if string(data[i]) == "1" {
			elevator.UpQueue[i] = 1
			elevator.DownQueue[i] = 1
		} else {
			elevator.UpQueue[i] = 0
			elevator.DownQueue[i] = 0
		}
	}
	file.Close()
}
