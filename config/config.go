package config

import "time"

const (
	Port = 16569

	NumberOfElevators = 3
	NumberOfFloors    = 4
	MaxCost           = 999999999
	ElevatorId        = 2

	ElevatorAddress   = "localhost:15657"
	//ElevatorAddress   = "10.24.32.40:15657"

	/*Constants*/
	TRAVEL_TIME int = 2 
	DOOR_OPEN_TIME int = 3

	/*Timer durations*/
	DOOR_OPEN_TIMER  time.Duration = 3
	MAX_TRAVEL_TIME 	time.Duration = 4
	MAX_OBSTRUCTION_TIME  time.Duration = 9 
	MAX_DOOR_CLOSE_TRIES 				  = 3

)
