package types

import (
	"time"
)

const (
	NumFloors   	 int = 4
	DOOROPENTIME 	 time.Duration = 3
	PASSINGFLOORTIME time.Duration =3
)

type State int

const (
	IDLE     State = 0
	MOVING         = 1
	DOOROPEN       = 2
	MOTORSTOP      = 3
)

type Direction int
const (
	Up 	  Direction = 1
	Down 			= -1
	Stop			= 0
)


// Constants
const (
	NumberOfElevators = 3 // Need better implemantation (config fil?)
	NumberOfFloors    = 4 // also config?
	maxCost           = 999999999
	elevatorId        = 0
)


// Structures
type Order struct {
	Floor         int
	DirectionUp   bool
	DirectionDown bool
	CabOrder      bool
	Cost          [NumberOfElevators]int
	Status        int  // 0: No active order , 1: waiting for cost, 2: unconfirmed, 3: confirmed, 4: mine, 5: done
	TimedOut      bool // Time? or Id?
}


type Elevator struct {
	UpQueue      [NumFloors]int
	DownQueue    [NumFloors]int
	CurrentFloor int
	Direction    int
	Motorstop    bool
}

type FsmChannels struct {
	FloorReached   chan int
	NewOrder       chan Order
	Obstruction    chan bool

	ElevatorState  chan Elevator

	DoorTimedOut   chan bool
	MotorTimedOut  chan int
	StopMotorTimer chan bool
}

