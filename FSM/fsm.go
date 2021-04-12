package fsm

import (
	"../elevio"
	"fmt"
	. "../types"
	. "../timer"
)

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

func runElevator(channels FsmChannels, OrderUpdate chan<- Order, ElevState chan<- Elevator) {
	State := IDLE
	var elevatorInfo channels.ElevatorState
	QueueDirection := Stop
	elevator.CurrentFloor = 0

	var nextFloor int
	var obstructed bool

	go elevio.PollFloorSensor(channels.FloorReached)
	go elevio.PollObstructionSwitch(channels.Obstruction)
	//go elevio.PollStopButton(channels.Stop)

	//for select switch case
	for{
		select{
		case newOrder := <- channels.NewOrder:
			fmt.println("New order to floor: ", NewOrder.Floor)

			switch State{
			case IDLE:
				//sjekk om du er i den etasjen fra før av
				if elevator.CurrentFloor == newOrder.Floor{

					elevio.SetDoorOpenLamp(true)
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut) 
					fmt.Println("Started Doortimer")

					State = DOOROPEN

				} else {
				
					//legger til i køen
					if newOrder.DirectionUp == true |  newOrder.CabOrder == true{
						elevatorInfo.UpQueue[newOrder.Floor] = 1
					}
					if newOrder.DirectionDown == true |  newOrder.CabOrder == true{
						elevatorInfo.DownQueue[newOrder.Floor] = 1
					}

					nextFloor := queueSearch(QueueDirection, elevatorInfo)
					dir := getDirection(elevator.CurrentFloor, nextFloor)
					elevio.SetMotorDirection(dir)
					QueueDirection = dir
					elevatorInfo.Direction = dir	

					
					//Start MotorStopTimer 
					go timer.StoppableTimer(PASSINGFLOORTIME, 1, channels.StopMotorTimer, channels.MotorTimedOut)
					fmt.Println("Started motortimer")
					State = MOVING

					//update elev-info
					ElevState <- elevatorInfo
				}
			case MOVING:
					//legger til i køen
				if newOrder.DirectionUp == true | newOrder.CabOrder == true{
					elevatorInfo.UpQueue[newOrder.Floor] = 1
				}
				if newOrder.DirectionDown == true | newOrder.CabOrder == true{ {
					elevatorInfo.DownQueue[newOrder.Floor] = 1
				}

				//Needs to be able 
				nextFloor := queueSearch(QueueDirection, elevatorInfo)

				//update elev-info
				ElevState <- elevatorInfo	
					
			case DOOROPEN:
				if elevator.CurrentFloor == newOrder.Floor{

					elevio.SetDoorOpenLamp(true)
					CountDownTimer(DOOROPENTIME, channels.DoorTimedOut) 
					fmt.Println("Started Doortimer")
					removeFromQueue(elevatorInfo.CurrentFloor)

					//send a completed order message to OrderDistributed
					NewOrder.Status = Done
					OrderUpdate <- NewOrder

				} else {
					//legger til i køen
					if newOrder.DirectionUp == true | newOrder.CabOrder == true {
						elevatorInfo.UpQueue[newOrder.Floor] = 1
					}
					if newOrder.DirectionDown == true | newOrder.CabOrder == true{
						elevatorInfo.DownQueue[newOrder.Floor] = 1
					}
					//update elev-info
					ElevState <- elevatorInfo					
				}
			case MOTORSTOP:
			}



		case floorArrival := <- channels.FloorReached:
			fmt.println("Arriving at floor: ", floorArrival)
			elevatorInfo.CurrentFloor = floorArrival
			SetFloorIndicator(floorArrival)
			elevatorInfo.CurrentFloor = floorArrival

			switch State{
			case IDLE:
			case MOVING:
				if nextFloor == floorArrival{
					elevio.SetMotorDirection(MD_Stop)
					elevatorInfo.Direction = Stop

					//Stop motorTimer
				    channels.StopMotorTimer <- true
					fmt.println("Started Motortimer")

					//open door
					SetDoorOpenLamp(true)
					removeFromQueue(elevatorInfo.CurrentFloor)
					
					//send a completed order message to OrderDistributed
					var Expidized_order Order
					Expidized_order.Floor = floorArrival
					Expidized_order.Status = Done
					OrderUpdate <- Expidized_order

					//starte door-timer
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.println("Started doortimer")
					State = DOOROPEN

					//update elev-info
					ElevState <- elevatorInfo
				} else {
					//Restart motorTimer
					channels.StopMotorTimer <- true
					go timer.StoppableTimer(PASSINGFLOORTIME, 1, channels.StopMotorTimer, channels.MotorTimedOut)
					fmt.println("Restarted Motortimer")
				}
			case DOOROPEN:
			case MOTORSTOP:
				//Kommer den hit, da har den ikke motorstopp lenger, men den har heller ikke noe i køen
				State = IDLE
				
				//si at jeg ikke har motorstopp lenger til Jon, da sender jon mine cabbies
				elevatorInfo.Motorstop = false
				ElevState <- elevatorInfo
			}



		
		case closeDoor := <- channels.DoorTimedOut:
			fmt.println("Door has timedout")
			switch State{
			case IDLE:
			case MOVING:
			case DOOROPEN:

				if obstructed == true{
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.println("Started doortimer")
					break
				}

			
				if checkOrdersPresent(elevatorInfo) == true{
					nextFloor := queueSearch(QueueDirection, elevatorInfo)
					dir = getDirection(elevatorInfo.currentFloor, nextFloor)
					elevio.SetMotorDirection(dir)
					QueueDirection = dir
					elevatorInfo.Direction = dir

					//start motor-timer
					go timer.StoppableTimer(PASSINGFLOORTIME, 1, channels.StopMotorTimer, channels.MotorTimedOut)
					fmt.Println("Started motortimer")
					State = MOVING

				} else {	
					State = IDLE
				}
				//update elev-info
				ElevState <- elevatorInfo

			case MOTORSTOP:
			}


		case obstructed = <- chanels.Obstruction:

		case motorStop := <- channels.MotorTimedOut:
			fmt.println("Motorstop detected")

			//tell OrderDistributer that I have motorstop 
			elevatorInfo.Motorstop = true

			//tømme min egen kø, jon sender den til andre heiser
			emptyQueue(elevatorInfo)

			//update elevInfo
			ElevState <- elevatorInfo
			State = MOTORSTOP


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

func checkOrdersPresent(elevator Elevator) {
	foundOrder := false
	for i := 1; i < NumFloors; i++ {
		if elevator.UpQueue[i] || elevator.DownQueue[i] == 1 {
			foundOrder = true
		}
	}
	return foundOrder
}

func queueSearch(QueueDirection int, elevator Elevator) {
	nextFloor := 0

	//first time
	if QueueDirection == Stop{
		QueueDirection = Up
	}


	if QueueDirection == Up {
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
	if QueueDirection == Down {
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

func removeFromQueue(Floor int){
	if elevator.Direction == Up {
		elevator.UpQueue[Floor] = 0
	}else{
		elevator.DownQueue[Floor] = 0
	}
}

func emptyQueue(elevatorInfo Elevator){
	for floor := 0; floor < NumFloors; floor++ {
		elevatorInfo.UpQueue[floor]=0
		elevatorInfo.DownQueue[floor]=0
	}
}
