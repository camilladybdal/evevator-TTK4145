package types


import (
	"../elevio"
   . "../config"
)

type State int

const (
	IDLE     State = 0
	MOVING         = 1
	DOOROPEN       = 2
	IMMOBILE      = 3
)

type Status int

const (
	NoActiveOrder  Status = 0 //burde byttes til NotActive
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
	Timestamp	  int64
}

type Elevator struct {
	UpQueue      [NumberOfFloors ]int
	DownQueue    [NumberOfFloors ]int
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

