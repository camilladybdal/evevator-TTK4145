package fsm

//import queue
//import timer-module
//import elevio

import (
	"fmt"

	"../elevio"
)

type state int

const (
	IDLE     state = 0
	MOVING         = 1
	DOOROPEN       = 2
)

type FsmChannels struct {
	//ButtonPress    chan elevio.ButtonEvent
	FloorReached   chan int
	MotorDirection chan int
	NewOrder       chan orderDistributer.Order
}

//hvordan lagrer vi numfloors som global var?
var upQueue [4]int
var downQueue [4]int

// ------------------------------------------------------------------------------------------------------------------------------------------------------
// Global functions
// ------------------------------------------------------------------------------------------------------------------------------------------------------

func InitFSM(numFloors int) {

	//go to first floor
	elevio.SetMotorDirection(elevio.MD_Down)
	for elevio.GetFloor() != 0 {
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetFloorIndicator(0)

	fmt.Println("FSM Initialized")

}

func runElevator(channels FsmChannels) {
	state := IDLE
	currentFloor := 0
	currentOrder orderDistributer.Order

	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollFloorSensor(channels.FloorReached)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		select {
		case IDLE:
			select {
			case currentOrder = <-channels.NewOrder:
				channels.MotorDirection <- getDirection(currentFloor, currentOrder.floor)

				//endre state til MOVING
				state = MOVING

				//update elevator info (?) hva er egt tanken her
				break
			}
		
		case MOVING:
			select{
			case direction := <- channels.MotorDirection:
				elevio.SetMotorDirection(direction)

				if direction == 0 {
					state = IDLE
				}

			case floor <- channels.FloorReached:
				currentFloor = floor
				elevio.SetFloorIndicator(floor) //hvordan skru av?

				if currentFloor == currentOrder.floor{
					elevio.SetMotorDirection(elevio.MD_Stop)
					state = DOOROPEN
				}	
			}

		case DOOROPEN:
			//open the door, start door-timer
			//after door-timer, close the door. 
			//if obcstruction: keep door open
			elevio.SetDoorOpenLamp(true)

			TimedOut := make(chan bool)
			go timer.DoorTimer(3, TimedOut)

			if <-TimedOut {
				elevio.SetDoorOpenLamp(false)
				state := IDLE
			}


			if <- drv_obstr{
				elevio.SetDoorOpenLamp(true)
				go timer.DoorTimer(3,TimedOut)  //er dette lov a?
			}

		}

		/*
		case <- drv_stop
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetStopLamp(true)
		*/

	}

}

// ------------------------------------------------------------------------------------------------------------------------------------------------------
// Local functions
// ------------------------------------------------------------------------------------------------------------------------------------------------------


func getDirection(currentFloor int, destinationFloor int ){
	if currentFloor - destinationFloor > 0{
		return elevio.MD_Down
	}
	else{
		return elevio.MD_Up
	}
}