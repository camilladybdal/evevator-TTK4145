package types

const (
	NumFloors    int = 4
	DOOROPENTIME int = 3
)

type State int

const (
	IDLE     State = 0
	MOVING         = 1
	DOOROPEN       = 2
)

type Elevator struct {
	UpQueue      [NumFloors]int
	DownQueue    [NumFloors]int
	CurrentFloor int
	Direction    int
}

// Constants
const (
	NumberOfElevators = 3 // Need better implemantation (config fil?)
	NumberOfFloors    = 4 // also config?
	MaxCost           = 999999999
	ElevatorId        = 0
)

type Status int
const (
	NoActiveOrder Status = 0
	WaitingForCost		 = 1
	Unconfirmed 		 = 2
	Confirmed 			 = 3
	Mine 				 = 4
	Done				 = 5
)

// Structures
type Order struct {
	Floor         int
	DirectionUp   bool
	DirectionDown bool
	Cost          [NumberOfElevators]int
	Status        Status  // 0: No active order , 1: waiting for cost, 2: unconfirmed, 3: confirmed, 4: mine, 5: done
	TimedOut      bool // Time? or Id?
}

// Button struct?