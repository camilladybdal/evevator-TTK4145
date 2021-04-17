package config

import "time"

const (
	Port = 16569

	NumberOfElevators = 2
	NumberOfFloors    = 4
	MaxCost           = 999999999
	ElevatorId        = 1

	ElevatorAddress   = "localhost:15657"
	//ElevatorAddress   = "10.24.1.24:15657"

	/*Constants*/
	TRAVEL_TIME int = 2 
	DOOR_OPEN_TIME int = 3

	/*Timer durations*/
	DOOR_OPEN_TIMER  time.Duration = 3
	MAX_TRAVEL_TIME 	time.Duration = 4
	MAX_OBSTRUCTION_TIME  time.Duration = 9 

)
