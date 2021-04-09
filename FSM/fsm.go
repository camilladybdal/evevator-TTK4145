package fsm

import (
	"../elevio"
	"fmt"
	. "../types"
)

type FsmChannels struct {
	FloorReached   chan int
	MotorDirection chan int
	NewOrder       chan Order
	Obstruction    chan bool
	Stop           chan bool
	ElevatorState  chan Elevator
}

//types.go

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
	var elevator Elevator
	var currentOrderFloor int
	var newOrder Order
	var QueueDirection int

	elevator.CurrentFloor = 0

	go elevio.PollFloorSensor(channels.FloorReached)
	go elevio.PollObstructionSwitch(channels.Obstruction)
	go elevio.PollStopButton(channels.Stop)

	for {
		switch State {
		case IDLE:
			select {
			case newOrder = <-channels.NewOrder:
				if newOrder.DirectionUp == true {
					elevator.UpQueue[newOrder.Floor] = 1
				}
				if newOrder.DirectionDown == true {
					elevator.DownQueue[newOrder.Floor] = 1
				}

			case checkOrdersPresent() == true:
				currentOrderFloor = queueSearch(QueueDirection)
				channels.MotorDirection <- getDirection(elevator.CurrentFloor, currentOrderFloor)
				State = MOVING
				break

				<-channels.Elevatorstate
				channels.Elevatorstate <- elevator
			}
		case MOVING:
			select {
			case elevator.Direction = <-channels.MotorDirection:
				elevio.SetMotorDirection(elevator.Direction)
				QueueDirection = elevator.Direction

				if Elevator.Direction == elevio.MD_Stop {
					State = IDLE
				}

			case elevator.CurrentFloor = <-channels.FloorReached:
				elevator.CurrentFloor = elevator.CurrentFloor
				elevio.SetFloorIndicator(elevator.CurrentFloor)

				if elevator.CurrentFloor == currentOrderFloor {
					elevio.SetMotorDirection(elevio.MD_Stop)
					State = DOOROPEN
				}
				<-channels.Elevatorstate
				channels.Elevatorstate <- elevator
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
			if <-channels.Obstruction {
				elevio.SetDoorOpenLamp(true)
				go timer.DoorTimer(DOOROPENTIME, TimedOut) //er dette lov a?
			}
			//rain TimedOut channel
			<-channels.Elevatorstate
			chanels.Elevatorstate <- elevator
			<-TimedOut
		}
	}
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------
// Local functions
// ------------------------------------------------------------------------------------------------------------------------------------------------------

func getDirection(currentFloor int, destinationFloor int) {
	if currentFloor-destinationFloor > 0 {
		return elevio.MD_Down
	} else {
		return elevio.MD_Up
	}
}

func checkOrdersPresent() {
	foundOrder := false
	for i := 1; i < NumFloors; i++ {
		if elevator.UpQueue[i] || elevator.DownQueue[i] == 1 {
			foundOrder = true
		}
	}
	return foundOrder
}

func queueSearch(QueueDirection int) {
	nextFloor := 0
	if QueueDirection == 1 {
		for floor := elevator.CurrentFloor; floor < NumFloors; floor++ {
			if elevator.UpQueue[floor] == 1 {
				nextFloor = elevator.UpQueue[floor]
				break
			}
		}
		for floor := NumFloors - 1; floor >= 0; floor-- {
			if elevator.DownQueue[floor] == 1 {
				nextFloor = elevator.DownQueue[floor]
				break
			}
		}
		for floor := 0; floor < Elevator.CurrentFloor; floor++ {
			if elevator.UpQueue[floor] == 1 {
				nextFloor = elevator.UpQueue[floor]
				break
			}
		}
	}
	if QueueDirection == -1 {
		for floor := elevator.CurrentFloor; floor >= 0; floor-- {
			if elevator.DownQueue[floor] == 1 {
				nextFloor = elevator.DownQueue[floor]
				break
			}
		}
		for floor := 0; floor < NumFloors; floor++ {
			if elevator.UpQueue[floor] == 1 {
				nextFloor = elevator.UpQueue[floor]
				break
			}
		}
		for floor := elevator.CurrentFloor; floor >= 0; floor-- {
			if elevator.DownQueue[floor] == 1 {
				nextFloor = elevator.DownQueue[floor]
				break
			}
		}
	}
	return nextFloor
}
