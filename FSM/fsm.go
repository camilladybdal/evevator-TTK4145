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
	var elevatorInfo Elevator
	elevatorInfo.CurrentFloor = 0
	var QueueDirection elevio.MotorDirection
	QueueDirection = elevio.MD_Stop

	var nextFloor int
	var obstructed bool

	go elevio.PollFloorSensor(channels.FloorReached)
	go elevio.PollObstructionSwitch(channels.Obstruction)


	//for select switch case
	for{
		select{
		case newOrder := <- channels.NewOrder:
			fmt.Println("New order to floor: ", newOrder.Floor)

			switch State{
			case IDLE:
				//sjekk om du er i den etasjen fra før av
				if elevatorInfo.CurrentFloor == newOrder.Floor{

					elevio.SetDoorOpenLamp(true)
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut) 
					fmt.Println("Started Doortimer")

					State = DOOROPEN

				} else {
				
					//legger til i køen
					if newOrder.DirectionUp == true ||  newOrder.CabOrder == true{
						elevatorInfo.UpQueue[newOrder.Floor] = 1
					}
					if newOrder.DirectionDown == true ||  newOrder.CabOrder == true{
						elevatorInfo.DownQueue[newOrder.Floor] = 1
					}

					nextFloor = queueSearch(QueueDirection, elevatorInfo)

					
					dir := getDirection(elevatorInfo.CurrentFloor, nextFloor)
					elevio.SetMotorDirection(dir)
					QueueDirection = dir
					elevatorInfo.Direction = dir	

					
					//Start IMMOBILETimer 
					go StoppableTimer(PASSINGFLOORTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("Started motortimer")
					State = MOVING

					//update elev-info
					ElevState <- elevatorInfo
				}
			case MOVING:
					//legger til i køen
				if newOrder.DirectionUp == true || newOrder.CabOrder == true{
					elevatorInfo.UpQueue[newOrder.Floor] = 1
				}
				if newOrder.DirectionDown == true || newOrder.CabOrder == true{ 
					elevatorInfo.DownQueue[newOrder.Floor] = 1
				}

				//Needs to be able 
				nextFloor = queueSearch(QueueDirection, elevatorInfo)

				//update elev-info
				ElevState <- elevatorInfo	
					
			case DOOROPEN:
				if elevatorInfo.CurrentFloor == newOrder.Floor{

					elevio.SetDoorOpenLamp(true)
					CountDownTimer(DOOROPENTIME, channels.DoorTimedOut) 
					fmt.Println("Started Doortimer")
					removeFromQueue(elevatorInfo)

					//send a completed order message to OrderDistributed
					newOrder.Status = 5 //replace with Done
					OrderUpdate <- newOrder

				} else {
					//legger til i køen
					if newOrder.DirectionUp == true || newOrder.CabOrder == true {
						elevatorInfo.UpQueue[newOrder.Floor] = 1
					}
					if newOrder.DirectionDown == true || newOrder.CabOrder == true{
						elevatorInfo.DownQueue[newOrder.Floor] = 1
					}
					//update elev-info
					ElevState <- elevatorInfo					
				}
			case IMMOBILE:
			}




		case floorArrival := <- channels.FloorReached:
			fmt.Println("Arriving at floor: ", floorArrival)
			elevatorInfo.CurrentFloor = floorArrival
			elevio.SetFloorIndicator(floorArrival)

			elevatorInfo.CurrentFloor = floorArrival

			switch State{
			case IDLE:
			case MOVING:
				if nextFloor == floorArrival{
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevatorInfo.Direction = elevio.MD_Stop

					//Stop motorTimer
				    channels.StopImmobileTimer <- true
					fmt.Println("Started Motortimer")

					//open door
					elevio.SetDoorOpenLamp(true)
					removeFromQueue(elevatorInfo)
					
					//send a completed order message to OrderDistributed
					var Expidized_order Order
					Expidized_order.Floor = floorArrival
					Expidized_order.Status = 5 //replace with Done
					OrderUpdate <- Expidized_order

					//starte door-timer
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.Println("Started doortimer")
					State = DOOROPEN

					//update elev-info
					ElevState <- elevatorInfo
				} else {
					//Restart motorTimer
					channels.StopImmobileTimer <- true
					go StoppableTimer(PASSINGFLOORTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("Restarted Motortimer")
				}
			case DOOROPEN:
			case IMMOBILE:
				//Kommer den hit, da har den ikke motorstopp lenger, men den har heller ikke noe i køen
				State = IDLE
				
				//si at jeg ikke har IMMOBILEp lenger til Jon, da sender jon mine cabbies
				elevatorInfo.Immobile = false
				ElevState <- elevatorInfo
			}

		
		case <- channels.DoorTimedOut:
			fmt.Println("Door has timedout")
			switch State{
			case IDLE:
			case MOVING:
			case DOOROPEN:

				if obstructed == true{
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.Println("Started doortimer")

					//starte obstruction timer 
					go StoppableTimer(MAXOBSTRUCTIONTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					break

				} else {
					//stop obstruction timer
					channels.StopImmobileTimer <- true
					elevio.SetDoorOpenLamp(false) //?
				}


				if checkOrdersPresent(elevatorInfo) == true{
					nextFloor := queueSearch(QueueDirection, elevatorInfo)
					dir := getDirection(elevatorInfo.CurrentFloor, nextFloor)
					elevio.SetMotorDirection(dir)
					QueueDirection = dir
					elevatorInfo.Direction = dir

					//start motor-timer
					go StoppableTimer(PASSINGFLOORTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("Started motortimer")
					State = MOVING

				} else {	
					State = IDLE
				}
				//update elev-info
				ElevState <- elevatorInfo

			case IMMOBILE:
				//hvis den er i denne staten pga obstruction, kommer den ut av den hvis døren lukker seg
				State = IDLE
				elevatorInfo.Immobile = false
				ElevState <- elevatorInfo
			}


		case obstructed = <- channels.Obstruction:

		case <- channels.Immobile:
			fmt.Println( "IMMOBILITY detected")

			//tell OrderDistributer that I am IMMOBILE 
			elevatorInfo.Immobile = true

			//tømme min egen kø, jon sender den til andre heiser
			emptyQueue(elevatorInfo)

			//update elevInfo
			ElevState <- elevatorInfo
			State = IMMOBILE

		}
	}
}


// ------------------------------------------------------------------------------------------------------------------------------------------------------
// Local functions
// ------------------------------------------------------------------------------------------------------------------------------------------------------

func getDirection(currentFloor int, destinationFloor int) elevio.MotorDirection {
	if currentFloor-destinationFloor > 0 {
		return elevio.MD_Down
	} else {
		return elevio.MD_Up
	}
}

func checkOrdersPresent(elevator Elevator) bool{
	foundOrder := false
	for i := 1; i < NumFloors; i++ {
		if elevator.UpQueue[i] ==1 || elevator.DownQueue[i] == 1 {
			foundOrder = true
		}
	}
	return foundOrder
}

func queueSearch(QueueDirection elevio.MotorDirection, elevator Elevator) int {
	nextFloor := 0

	//first time
	if QueueDirection == elevio.MD_Stop{
		QueueDirection = elevio.MD_Up
	}


	if QueueDirection == elevio.MD_Up {
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
		for floor := 0; floor < elevator.CurrentFloor; floor++ {
			if elevator.UpQueue[floor] == 1 {
				nextFloor = elevator.UpQueue[floor]
				break
			}
		}
	}
	if QueueDirection == elevio.MD_Down{
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

func removeFromQueue(elevator Elevator){
		elevator.UpQueue[elevator.CurrentFloor] = 0
		elevator.DownQueue[elevator.CurrentFloor] = 0
}

func emptyQueue(elevator Elevator){
	for floor := 0; floor < NumFloors; floor++ {
		elevator.UpQueue[floor] = 0
		elevator.DownQueue[floor] = 0
	}
}
