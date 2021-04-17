package costfnc

import (
	"fmt"

	. "../elevio"
	. "../types"
	. "../config"
)

//Based on minimal waiting time
func Costfunction(elev Elevator, neworder Order) int {
	cost := 0
	lastFloor := elev.CurrentFloor
	movement := getMovementList(elev, neworder)
	fmt.Println("Movement: ", movement)
	for i := 0; i < len(movement); i++ {
		cost += DOOR_OPEN_TIME
		cost += TRAVEL_TIME * abs(lastFloor-movement[i])
		lastFloor = movement[i]
		fmt.Println("Last_floor: ", lastFloor)
		fmt.Println("Current floor: ", movement[i])
		fmt.Println("Cost: ", cost)
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

//Only know queuedir
func getMovementList(elev Elevator, neworder Order) []int {
	var movement []int
	if elev.Direction == MD_Down {
		for i := elev.CurrentFloor; i >= 0; i-- { //Search downqueue for orders from current to bottom, if order in downqueue: + DOOROPENTIME + TRAVELTIME
			if i == neworder.Floor && neworder.DirectionDown == true {
				movement = append(movement, neworder.Floor)
				return movement
			}
			if elev.DownQueue[i] == 1 {
				movement = append(movement, i)
			}
		}
		for i := 0; i < NumberOfFloors; i++ { //Search upqueue for orders, if order in upqueue: + DOOROPENTIME + TRAVELTIME
			if i == neworder.Floor && neworder.DirectionUp == true {
				movement = append(movement, neworder.Floor)
				return movement
			}
			if elev.UpQueue[i] == 1 && elev.DownQueue[i] != 1 {
				movement = append(movement, i)
			}
		}
		for i := NumberOfFloors - 1; i < elev.CurrentFloor; i-- { //Search downqueue for orders from top to current
			if i == neworder.Floor && neworder.DirectionDown == true {
				movement = append(movement, neworder.Floor)
				return movement
			}
			if elev.DownQueue[i] == 1 {
				movement = append(movement, i)
			}
		}
	} else if elev.Direction == MD_Up {
		for i := elev.CurrentFloor; i < NumberOfFloors; i++ { //Search upqueue for orders from current floor
			if i == neworder.Floor && neworder.DirectionUp == true {
				movement = append(movement, neworder.Floor)
				return movement
			}
			if elev.UpQueue[i] == 1 {
				movement = append(movement, i)
			}
		}
		for i := NumberOfFloors - 1; i >= 0; i-- { //Search downqueue for orders
			if i == neworder.Floor && neworder.DirectionDown == true {
				movement = append(movement, neworder.Floor)
				return movement
			}
			if elev.DownQueue[i] == 1 && elev.UpQueue[i] != 1 { //Assume that everyone leaves/enters at a floor
				movement = append(movement, i)
			}
		}
		for i := 0; i < elev.CurrentFloor; i++ { //Search upqueue for orders from bottom to current
			if i == neworder.Floor && neworder.DirectionUp == true {
				movement = append(movement, neworder.Floor)
				return movement
			}
			if elev.UpQueue[i] == 1 { //Assume everyone leaves/enters elev at a floor
				movement = append(movement, i)
			}
		}
	}
	//What if elev has direction stop?
	movement = append(movement, neworder.Floor)
	return movement
}
