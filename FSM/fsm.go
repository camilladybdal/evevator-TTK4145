package fsm

//import queue
//import timer-module
//import elevio

import (
	"./elevio"
)

type State int

const (
	IDLE State = 0
	MOVING   = 1
	DOOROPEN = 2
)

type FsmChannels struct {
	ButtonPress    chan elevio.ButtonEvent
	FloorReached   chan int
	MotorDirection chan int
	NewOrder       chan orderDistributer.Order
}



var upQueue [floor]int
var downQueue [floor]int


/*func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := getFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}*/


func initFSM() {
	
	if elevio.getFloor() == -1{
		elevio.
	}


	State := IDLE

}

func runElevator() {}



