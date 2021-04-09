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
	maxCost           = 999999999
	elevatorId        = 0
)

type Status int
const (
	noActiveOrder Status = 0
	waitingForCost		 = 1
	unconfirmed 		 = 2
	confirmed 			 = 3
	mine 				 = 4
	done				 = 5
)

// Structures
type Order struct {
	Floor         int
	DirectionUp   bool
	DirectionDown bool
	Cost          [NumberOfElevators]int
	Status        int  // 0: No active order , 1: waiting for cost, 2: unconfirmed, 3: confirmed, 4: mine, 5: done
	TimedOut      bool // Time? or Id?
}

// Button struct?