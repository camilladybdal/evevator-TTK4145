package config

import "time"

const (
	/* Configurable */
	PORT                = 16569
	NUMBER_OF_ELEVATORS = 3
	NUMBER_OF_FLOORS    = 4
	ELEVATOR_ID         = 1
	ELEVATOR_ADDRESS    = "localhost:15657"

	/*Constants*/
	MAXCOST              int = 999999999
	TRAVEL_TIME          int = 2
	DOOR_OPEN_TIME       int = 3
	MAX_DOOR_CLOSE_TRIES int = 3

	/*Timer durations*/
	DOOR_OPEN_TIME_DURATION time.Duration = 3
	MAX_TRAVEL_TIME         time.Duration = 4
	MAX_OBSTRUCTION_TIME    time.Duration = 9
)
