package costfunction

import (
	. "../config"
	. "../elevio"
	. "../types"
)

func Costfunction(elevator Elevator, neworder Order) int {
	cost := 0
	lastFloor := elevator.CurrentFloor
	movementList := getMovementList(elevator, neworder)
	for i := 0; i < len(movementList); i++ {
		cost += DOOR_OPEN_TIME
		cost += TRAVEL_TIME * abs(lastFloor-movementList[i])
		lastFloor = movementList[i]
	}
	return cost
}

func abs(value int) int {
	if value < 0 {
		value = -value
		return value
	}
	return value
}

func getMovementList(elevator Elevator, neworder Order) []int {
	var movementList []int
	if elevator.Direction == MD_Down {
		for i := elevator.CurrentFloor; i >= 0; i-- {
			if i == neworder.Floor && neworder.DirectionDown == true {
				movementList = append(movementList, neworder.Floor)
				return movementList
			}
			if elevator.DownQueue[i] == 1 {
				movementList = append(movementList, i)
			}
		}
		for i := 0; i < NUMBER_OF_FLOORS; i++ {
			if i == neworder.Floor && neworder.DirectionUp == true {
				movementList = append(movementList, neworder.Floor)
				return movementList
			}
			if elevator.UpQueue[i] == 1 && elevator.DownQueue[i] != 1 {
				movementList = append(movementList, i)
			}
		}
		for i := NUMBER_OF_FLOORS - 1; i < elevator.CurrentFloor; i-- {
			if i == neworder.Floor && neworder.DirectionDown == true {
				movementList = append(movementList, neworder.Floor)
				return movementList
			}
			if elevator.DownQueue[i] == 1 {
				movementList = append(movementList, i)
			}
		}
	} else if elevator.Direction == MD_Up {
		for i := elevator.CurrentFloor; i < NUMBER_OF_FLOORS; i++ {
			if i == neworder.Floor && neworder.DirectionUp == true {
				movementList = append(movementList, neworder.Floor)
				return movementList
			}
			if elevator.UpQueue[i] == 1 {
				movementList = append(movementList, i)
			}
		}
		for i := NUMBER_OF_FLOORS - 1; i >= 0; i-- {
			if i == neworder.Floor && neworder.DirectionDown == true {
				movementList = append(movementList, neworder.Floor)
				return movementList
			}
			if elevator.DownQueue[i] == 1 && elevator.UpQueue[i] != 1 {
				movementList = append(movementList, i)
			}
		}
		for i := 0; i < elevator.CurrentFloor; i++ {
			if i == neworder.Floor && neworder.DirectionUp == true {
				movementList = append(movementList, neworder.Floor)
				return movementList
			}
			if elevator.UpQueue[i] == 1 {
				movementList = append(movementList, i)
			}
		}
	}
	movementList = append(movementList, neworder.Floor)
	return movementList
}
