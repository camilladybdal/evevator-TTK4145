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
	var elevatorInfo Elevator
	//var currentOrderFloor int
	var QueueDirection int

	elevator.CurrentFloor = 0

	go elevio.PollFloorSensor(channels.FloorReached)
	go elevio.PollObstructionSwitch(channels.Obstruction)
	go elevio.PollStopButton(channels.Stop)

	//for select switch case
	for{
		select{
		case newOrder := <- channels.NewOrder:

			//legger til i køen
			if newOrder.DirectionUp == true {
				elevatorInfo.UpQueue[newOrder.Floor] = 1
			}
			if newOrder.DirectionDown == true {
				elevatorInfo.DownQueue[newOrder.Floor] = 1
			}

			switch State{
			case IDLE:
				
				

			}
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
