package costfnc

import (
	"fmt"

	. "../elevio"
	. "../types"
)

//Based on minimal waiting time

var TRAVEL_TIME int = 2 //Perhaps change this number
var DOOR_OPEN_TIME int = 3

/*
func Costfunction(elev Elevator, neworder Order) int {
	cost := (elev.CurrentFloor - neworder.Floor) * TRAVEL_TIME
	if cost == 0 {
		return cost
	}
	if cost > 0 {
		if elev.Direction == -1 {
			for i := neworder.Floor; i < elev.CurrentFloor; i++ {
				if elev.DownQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
				}
			}
			return cost
		} else {
			cost = 0
			highestfloorabove := 0
			lowest_floor_below := 0
			for i := elev.CurrentFloor; i < NumFloors; i++ { //Search upqueue for orders from current floor, if order in upqueue: + DOOROPENTIME
				if elev.UpQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
					highestfloorabove = i
				}
				fmt.Println("Cost after iterating from current to top: ", cost)
			}
			for i := NumFloors - 1; i >= 0; i-- { //Search downqueue for orders, if order in downqueue: + DOOROPENTIME
				if elev.DownQueue[i] == 1 && elev.UpQueue[i] != 1 { //Assume that everyone leaves/enters at a floor
					cost += DOOR_OPEN_TIME
					if i > highestfloorabove {
						highestfloorabove = i
					}
					lowest_floor_below = i
				}
				fmt.Println("Cost after iterating from top to bottom: ", cost)
			}
			for i := 0; i < elev.CurrentFloor; i++ { //Search upqueue for orders from bottom to current, if order in upqueue: + DOOROPENTIME
				if elev.UpQueue[i] == 1 && elev.DownQueue[i] != 1 { //Assume everyone leaves/enters elev at a floor
					cost += DOOR_OPEN_TIME
					if i < lowest_floor_below {
						lowest_floor_below = i
					}
				}
				fmt.Println("Cost after iterating from bottom to current: ", cost)
			}
			if neworder.Floor < lowest_floor_below {
				lowest_floor_below = neworder.Floor
			}
			fmt.Println("highestfloorabove: ", highestfloorabove)
			fmt.Println("lowstfloorbelow: ", lowest_floor_below)
			if neworder.Floor == lowest_floor_below {
				cost += TRAVEL_TIME*(highestfloorabove-elev.CurrentFloor) + TRAVEL_TIME*(highestfloorabove-lowest_floor_below)
			} else {
				cost += TRAVEL_TIME*(highestfloorabove) + TRAVEL_TIME*(highestfloorabove-lowest_floor_below-1)
			}
			return cost
		}
	}
	if cost < 0 {
		cost = -cost
		if elev.Direction == 1 {
			for i := elev.CurrentFloor; i < neworder.Floor; i++ {
				if elev.UpQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
				}
			}
			return cost
		} else {
			cost = 0
			highestfloorabove := 0
			lowest_floor_below := 0
			for i := elev.CurrentFloor; i >= 0; i-- { //Search downqueue for orders from current to bottom, if order in downqueue: + DOOROPENTIME + TRAVELTIME
				if elev.DownQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
					lowest_floor_below = i
				}

			}
			for i := 0; i < NumFloors; i++ { //Search upqueue for orders, if order in upqueue: + DOOROPENTIME + TRAVELTIME
				if elev.UpQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
					if i != NumFloors-1 {
						cost += TRAVEL_TIME
					}
				}
			}
			for i := NumFloors - 1; i < elev.CurrentFloor; i-- { //Search downqueue for orders from top to current, if orders in downqueue: + DOORTIME + TRAVELTIME
				if elev.DownQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
					cost += TRAVEL_TIME
				}
			}
			return cost
		}
	}
	return cost
}*/

//Based on minimal movement
/*
func costfunction(neworder <-chan Order, localelevstate){
	cost := localelevstate.floor - order.floor
	if cost == 0 {
		return cost //(broadcast dette)
	}
	if cost > 0 {
		if localelevstate.motordir != -1 {
			cost = cost + numFloors
		}
	}
	if cost < 0 {
		cost = abs(cost)
		if localelevstate.motordir != 1 {
			cost = cost + numFloors
		}
	}
	return cost //(broadcast dette)

}
*/

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
		for i := 0; i < NumFloors; i++ { //Search upqueue for orders, if order in upqueue: + DOOROPENTIME + TRAVELTIME
			if i == neworder.Floor && neworder.DirectionUp == true {
				movement = append(movement, neworder.Floor)
				return movement
			}
			if elev.UpQueue[i] == 1 && elev.DownQueue[i] != 1 {
				movement = append(movement, i)
			}
		}
		for i := NumFloors - 1; i < elev.CurrentFloor; i-- { //Search downqueue for orders from top to current
			if i == neworder.Floor && neworder.DirectionDown == true {
				movement = append(movement, neworder.Floor)
				return movement
			}
			if elev.DownQueue[i] == 1 {
				movement = append(movement, i)
			}
		}
	} else if elev.Direction == MD_Up {
		for i := elev.CurrentFloor; i < NumFloors; i++ { //Search upqueue for orders from current floor
			if i == neworder.Floor && neworder.DirectionUp == true {
				movement = append(movement, neworder.Floor)
				return movement
			}
			if elev.UpQueue[i] == 1 {
				movement = append(movement, i)
			}
		}
		for i := NumFloors - 1; i >= 0; i-- { //Search downqueue for orders
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
