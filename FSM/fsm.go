package fsm

import (
	"../elevio"
	"fmt"
	. "../types"
	. "../timer"
	//"time"
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

	elevio.SetDoorOpenLamp(false)

	fmt.Println("FSM Initialized")
}

func RunElevator(channels FsmChannels, OrderUpdate chan<- Order, ElevState chan<- Elevator) {
	State := IDLE
	var elevatorInfo Elevator
	emptyQueue(&elevatorInfo)

	elevatorInfo.CurrentFloor = 0
	wasobstr := false
	
	var QueueDirection elevio.MotorDirection 
	QueueDirection = elevio.MD_Stop

	var nextFloor int
	var obstructed bool

	go elevio.PollFloorSensor(channels.FloorReached)
	go elevio.PollObstructionSwitch(channels.Obstruction)
	fmt.Println("Polling started...")

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

					//det er her det går galt
					nextFloor = queueSearch(QueueDirection, elevatorInfo)
					fmt.Println("floor im heading for is: ", nextFloor)

					
					dir := getDirection(elevatorInfo.CurrentFloor, nextFloor)
					elevio.SetMotorDirection(dir)
					
					QueueDirection = dir
					elevatorInfo.Direction = dir	
					fmt.Println("direction to floor is: " , dir)

					
					//Start for motor!!
					go StoppableTimer(PASSINGFLOORTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("Started motortimer")
					State = MOVING

					//update elev-info
					//ElevState <- elevatorInfo
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
				//ElevState <- elevatorInfo	
					
			case DOOROPEN:
				if elevatorInfo.CurrentFloor == newOrder.Floor{

					elevio.SetDoorOpenLamp(true)
					CountDownTimer(DOOROPENTIME, channels.DoorTimedOut) 
					fmt.Println("Started Doortimer")
					removeFromQueue(&elevatorInfo)

					//send a completed order message to OrderDistributed
					newOrder.Status = 5 //replace with Done
					//OrderUpdate <- newOrder

				} else {
					//legger til i køen
					if newOrder.DirectionUp == true || newOrder.CabOrder == true {
						elevatorInfo.UpQueue[newOrder.Floor] = 1
					}
					if newOrder.DirectionDown == true || newOrder.CabOrder == true{
						elevatorInfo.DownQueue[newOrder.Floor] = 1
					}
					//update elev-info
					//ElevState <- elevatorInfo					
				}
			case IMMOBILE:
			}




		case floorArrival := <- channels.FloorReached:
			fmt.Println("Arriving at floor: ", floorArrival)
			elevatorInfo.CurrentFloor = floorArrival
			elevio.SetFloorIndicator(floorArrival)

			switch State{
			case IDLE:
			case MOVING:
				if nextFloor == floorArrival{

					elevio.SetMotorDirection(elevio.MD_Stop)
					elevatorInfo.Direction = elevio.MD_Stop


					//Stop motorTimer
				    channels.StopImmobileTimer <- true
					fmt.Println("Stopped Motortimer")
					fmt.Println("Floor im heading for is :" ,nextFloor)

					//open door
					elevio.SetDoorOpenLamp(true)					
					removeFromQueue(&elevatorInfo)
					
					//send a completed order message to OrderDistributed
					var Expidized_order Order
					Expidized_order.Floor = floorArrival
					Expidized_order.Status = 5 //replace with Done
					//OrderUpdate <- Expidized_order

					//starte door-timer
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.Println("Started doortimer")
					State = DOOROPEN

					//update elev-info
					//ElevState <- elevatorInfo
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
				//ElevState <- elevatorInfo
			}

		
		case <- channels.DoorTimedOut:
			fmt.Println("Door has timedout")
			fmt.Println("my state is : ", State)
			
			
			switch State{
			case IDLE:
			case MOVING:
			case DOOROPEN:
				

				if obstructed == true{
					fmt.Println("OBSTRUCTION")

					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.Println("Restarted doortimer")

					if wasobstr == false {
					//starte obstruction timer  første gangen den er obsruert
					fmt.Println("started obstruction timer")
					go StoppableTimer(MAXOBSTRUCTIONTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					}

					wasobstr = true //nå er denne forever true frem til hele driten omgangen er ferdig

				}

				
				//stop obstruction timer
				if obstructed == false && wasobstr == true{
					fmt.Println("OBSTRUCTION OFF")

					fmt.Println("Stopping immobility timer here 1 ")
					channels.StopImmobileTimer <- true
					elevio.SetDoorOpenLamp(false) 
					}
				

				if checkOrdersPresent(elevatorInfo) == true{
					elevio.SetDoorOpenLamp(false)
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
					if (obstructed == false && checkOrdersPresent(elevatorInfo) == false){
						elevio.SetDoorOpenLamp(false)
						State = IDLE
						fmt.Println("im here, have closed door, state idle")
					}
				}

				//update elev-info
				//ElevState <- elevatorInfo

			case IMMOBILE:
				
				fmt.Println("entered here")

				if obstructed == false {
					State = IDLE
					elevatorInfo.Immobile = false
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.Println("Restarted doortimer")

				}
				//ElevState <- elevatorInfo 
				//elevstaten har jo ikke endret seg nå...
			}


		case obstructed = <- channels.Obstruction:
				fmt.Println("Obstruction is : ", obstructed)

				if obstructed == false {
					if State == IMMOBILE{
						State = DOOROPEN
						//channels.StopImmobileTimer <- true
						go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
						fmt.Println("Restarted doortimer, no longer obstructed")
						wasobstr = false
					}
				}

		case <- channels.Immobile:
			fmt.Println( "IMMOBILITY detected")

			//stop immobility timer
			channels.StopImmobileTimer <- true

			//tell OrderDistributer that I am IMMOBILE 
			elevatorInfo.Immobile = true

			//tømme min egen kø, jon sender den til andre heiser
			emptyQueue(&elevatorInfo)

			//update elevInfo
			//ElevState <- elevatorInfo
			State = IMMOBILE


		default:
		}
	}
}



