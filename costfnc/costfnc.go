package costfnc

const (
	NumFloors int = 4
)

type Order struct {
	Floor int
}

type Elevator struct {
	UpQueue      [NumFloors]int
	DownQueue    [NumFloors]int
	CurrentFloor int
	Direction    int
}

//Based on minimal waiting time

var TRAVEL_TIME int = 2 //Perhaps change this number
var DOOR_OPEN_TIME int = 3

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
			for i := elev.CurrentFloor; i < NumFloors-1; i++ {
				if elev.UpQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
				}
			}
			for i := neworder.Floor; i < NumFloors-1; i++ {
				if elev.DownQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
				}
			}
			return cost
		}
	}
	if cost > 0 {
		cost = -cost
		if elev.Direction == 1 {
			for i := elev.CurrentFloor; i < neworder.Floor; i++ {
				if elev.UpQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
				}
			}
			return cost
		} else {
			for i := 0; i < elev.CurrentFloor; i++ {
				if elev.DownQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
				}
			}
			for i := 0; i < neworder.Floor; i++ {
				if elev.UpQueue[i] == 1 {
					cost += DOOR_OPEN_TIME
				}
			}
			return cost
		}
	}
	return cost
}

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
