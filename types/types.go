package types


import (
	"time"
	"../elevio"
   . "../config"
)

const (
	NumFloors   	 	int = 4
	DOOROPENTIME 	 	time.Duration = 3
	PASSINGFLOORTIME 	time.Duration = 4
	MAXOBSTRUCTIONTIME  time.Duration = 9 //?
)

type State int

const (
	IDLE     State = 0
	MOVING         = 1
	DOOROPEN       = 2
	IMMOBILE      = 3
)

/*
type Direction int
const (
	Up 	  Direction = 1
	Down 			= -1
	Stop			= 0
)
*/

// Constants

type Status int

const (
	NoActiveOrder  Status = 0
	WaitingForCost        = 1
	Unconfirmed           = 2
	Confirmed             = 3
	Mine                  = 4
	Done                  = 5
)


// Structures
type Order struct {
	Floor         int
	DirectionUp   bool
	DirectionDown bool
	CabOrder      bool
	Cost          [NumberOfElevators]int
	Status        Status // 0: No active order , 1: waiting for cost, 2: unconfirmed, 3: confirmed, 4: mine, 5: done
	TimedOut      bool   // Time? or Id?
	FromId		  int
}



type Elevator struct {
	UpQueue      [NumFloors]int
	DownQueue    [NumFloors]int
	CurrentFloor int
	Direction    elevio.MotorDirection
	Immobile    bool
}

type FsmChannels struct {
	FloorReached   chan int
	NewOrder       chan Order
	Obstruction    chan bool

	ElevatorState  chan Elevator

	DoorTimedOut   chan bool
	Immobile 		chan int
	StopImmobileTimer chan bool
}

