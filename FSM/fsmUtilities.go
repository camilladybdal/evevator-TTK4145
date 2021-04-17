package fsm

import (
	"fmt"
	"os"

	"bufio"
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

func openBackupFile() {
	if _, err := os.Stat("backupCabOrders.txt"); os.IsNotExist(err) {
		filename, err := os.Create("backupCabOrders.txt")
		checkError(err)
		defer filename.Close()
		for i := 0; i < NumberOfFloors; i++ {
			_, err1 := filename.WriteString(fmt.Sprintf("%d\n", 0))
			checkError(err1)
		}
	}
	filename, err := os.Open("backupCabOrders.txt")
	checkError(err)
	defer filename.Close()
}

func readFromBackupFile(filename *os.File) []int {
	var CabOrderArray []int
	scanner := bufio.NewScanner(filename)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		data, err := strconv.Atoi(scanner.Text())
		checkError(err)
		CabOrderArray = append(CabOrderArray, data)
	}
	return CabOrderArray
}

func writeToBackUpFile(filename *os.File) {
	//når du får orderen inn i fsm, så må du skrive ned til fil
	// dvs. skriv 1 til fil på den linja du har floor til
	//når du har utført orderen må du skrive 0 til linja med floor-nr
}
