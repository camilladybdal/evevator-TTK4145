package fsm

//import queue
//import timer-module
//import elevio

import (
	"fmt"

	"../elevio"
)

type State int

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
	Obstruction    chan bool
	Stop	       chan bool
}


type Elevator struct {
	UpQueue[NumFloors]   int
	DownQueue[NumFloors] int
	CurrentFloor 		 int
}

var DOOROPENTIME int = 3

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
	State := IDLE
	currentFloor := 0
	currentOrder orderDistributer.Order
	

	go elevio.PollFloorSensor(channels.FloorReached)
	go elevio.PollObstructionSwitch(channels.Obstruction)
	go elevio.PollStopButton(channels.Stop)

	for {
		select {
		case IDLE:
			select {
			case currentOrder = <-channels.NewOrder:
				channels.MotorDirection <- getDirection(currentFloor, currentOrder.floor)

				//add to down or up- queue
				

				//endre state til MOVING
				State = MOVING

				//update elevator info (?) hva er egt tanken her
				break
			}
		
		case MOVING:
			select{
			case direction := <- channels.MotorDirection:
				elevio.SetMotorDirection(direction)

				if direction == 0 {
					State = IDLE
				}

			case floor <- channels.FloorReached:
				currentFloor = floor
				elevio.SetFloorIndicator(floor) //hvordan skru av?

				if currentFloor == currentOrder.floor{
					elevio.SetMotorDirection(elevio.MD_Stop)
					State = DOOROPEN
				}	
			}

		case DOOROPEN:
			elevio.SetDoorOpenLamp(true)

			TimedOut := make(chan bool)
			go timer.DoorTimer(DOOROPENTIME, TimedOut)

			if <-TimedOut {
				elevio.SetDoorOpenLamp(false)
				State := IDLE
				break
			}


			if <- channels.Obstruction{
				elevio.SetDoorOpenLamp(true)
				go timer.DoorTimer(DOOROPENTIME,TimedOut)  //er dette lov a?
			}

			//drain TimedOut channel

		}

		/*
		case <- Stop
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