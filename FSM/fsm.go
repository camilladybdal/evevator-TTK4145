package fsm

//import queue
//import timer-module
//import elevio

type State int

const (
	IDLE     = 0
	MOVING   = 1
	DOOROPEN = 2
	TIMEDOUT = 3
)

/*function initFSM:
move down to floor 1
elev_io init with correct adress and port (can also be done in main)
*/

/* function runElevator:

state:= IDLE
floor = 0
current order = none


go elevio.PollButtons(drv_buttons)
go elevio.PollFloorSensor(drv_floors)
go elevio.PollObstructionSwitch(drv_obstr)
go elevio.PollStopButton(drv_stop)



look at state-diagram for rough draft on fsm-states and working

*/
