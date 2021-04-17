package fsm

import (
	"../elevio"
	"fmt"
	. "../types"
	. "../timer"
	. "../config"
	"time"
)

// ------------------------------------------------------------------------------------------------------------------------------------------------------
// Global functions
// ------------------------------------------------------------------------------------------------------------------------------------------------------

func InitFSM(numFloors int) {
	elevio.SetMotorDirection(elevio.MD_Down)
	for elevio.GetFloor() != 0 {
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetFloorIndicator(0)

	elevio.SetDoorOpenLamp(false)

	//initialize backup-file

	fmt.Println("FSM Initialized ")
}

func RunElevator(channels FsmChannels, OrderUpdate chan<- Order, ElevState chan<- Elevator) {
	State := IDLE
	var elevatorInfo Elevator
	emptyQueue(&elevatorInfo)

	elevatorInfo.CurrentFloor = 0
	wasobstr := false
	updateFileAndElevator:= false
	
	var QueueDirection elevio.MotorDirection 
	QueueDirection = elevio.MD_Stop
	
	//newCountdownTime <-chan time.Duration
	resetDoor := make(chan time.Duration)
	go ResetableTimer(time.Duration(0), resetDoor, channels.DoorTimedOut)
	<- channels.DoorTimedOut

	var nextFloor int
	var obstructed bool
	var immobilityNextFloor int

	//read cab-orders from file and add to queues. 
	readFromBackupFile("CabOrders", ElevatorId, &elevatorInfo)

	go elevio.PollFloorSensor(channels.FloorReached)
	go elevio.PollObstructionSwitch(channels.Obstruction)
	fmt.Println("Polling started...")

	//for select switch case
	for{
		select{
		case newOrder := <- channels.NewOrder:
			fmt.Println("----New order to floor: ", newOrder.Floor)
			fmt.Println("---- my State is: ", State)

			switch State{
			case IDLE:
				//sjekk om du er i den etasjen fra før av
				if elevatorInfo.CurrentFloor == newOrder.Floor{

					elevio.SetDoorOpenLamp(true)
					resetDoor <- DOOR_OPEN_TIMER
					
					
					fmt.Println("---- Started Doortimer")

					newOrder.Status = Done
					newOrder.FromId = ElevatorId
					OrderUpdate <- newOrder

					State = DOOROPEN

				} else {
				
					//legger til i køen
					if newOrder.DirectionUp == true ||  newOrder.CabOrder == true{
						elevatorInfo.UpQueue[newOrder.Floor] = 1
					}
					if newOrder.DirectionDown == true ||  newOrder.CabOrder == true{
						elevatorInfo.DownQueue[newOrder.Floor] = 1
					}

					//if caborder: skriv den til fil 


					//Søker etter floor HVIS den er i IDLE
					nextFloor = queueSearch(QueueDirection, elevatorInfo)
					fmt.Println("---- floor im heading for is: ", nextFloor)

					dir := getDirection(elevatorInfo.CurrentFloor, nextFloor)
					elevio.SetMotorDirection(dir)
					QueueDirection = dir
					elevatorInfo.Direction = dir	

					fmt.Println("---- direction to floor is: " , dir)
					
					//Start Motortimer
					go StoppableTimer(MAX_TRAVEL_TIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("---- Started motortimer")
					State = MOVING

					//update elev-info
					//ElevState <- elevatorInfo
					updateFileAndElevator = true
				}
			case MOVING:
				//legger til i køen
				if newOrder.DirectionUp == true || newOrder.CabOrder == true{
					elevatorInfo.UpQueue[newOrder.Floor] = 1
				}
				if newOrder.DirectionDown == true || newOrder.CabOrder == true{ 
					elevatorInfo.DownQueue[newOrder.Floor] = 1
				}

				//if caborder: skriv den til fil 


				if queueSearch(QueueDirection, elevatorInfo) == elevatorInfo.CurrentFloor{
					break
				} else{
					nextFloor = queueSearch(QueueDirection, elevatorInfo)
					fmt.Println("----my next floor is:", nextFloor)
				}

				//update elev-info
				//ElevState <- elevatorInfo
				updateFileAndElevator = true
	
					
			case DOOROPEN:
				if elevatorInfo.CurrentFloor == newOrder.Floor{

					elevio.SetDoorOpenLamp(true)

					//Reset this timer, dont start a new
					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println(" ---- Started Doortimer")
					removeFromQueue(&elevatorInfo)

					//send a completed order message to OrderDistributed
					newOrder.Status = Done
					newOrder.FromId = ElevatorId
					OrderUpdate <- newOrder

				} else {
					//legger til i køen
					if newOrder.DirectionUp == true || newOrder.CabOrder == true {
						elevatorInfo.UpQueue[newOrder.Floor] = 1
					}
					if newOrder.DirectionDown == true || newOrder.CabOrder == true{
						elevatorInfo.DownQueue[newOrder.Floor] = 1
					}

					//if caborder: skriv den til fil 

					//update elev-info
					//ElevState <- elevatorInfo	
					updateFileAndElevator = true
				
				}
			case IMMOBILE:
			}




		case floorArrival := <- channels.FloorReached:
			fmt.Println("---- Arriving at floor: ", floorArrival)

			elevatorInfo.CurrentFloor = floorArrival
			elevio.SetFloorIndicator(floorArrival)

			switch State{
			case IDLE:
			case MOVING:
				
				nextFloor = queueSearch(QueueDirection, elevatorInfo)
				fmt.Println("---- I am heding for this floor: ", nextFloor)
				
				//hvis den kommer her og har -1 har den vært immobil ?
				if nextFloor == -1 { 
					nextFloor = immobilityNextFloor
				}

				if nextFloor == floorArrival{

					elevio.SetMotorDirection(elevio.MD_Stop)
					elevatorInfo.Direction = elevio.MD_Stop


					//Stop motorTimer
				    channels.StopImmobileTimer <- true
					fmt.Println("---- Stopped Motortimer")

					//open door
					elevio.SetDoorOpenLamp(true)					
					removeFromQueue(&elevatorInfo)
				
					
					//send a completed order message to OrderDistributed
					expidizeOrder(elevatorInfo, OrderUpdate)

					//if caborder: skriv null til fil.


					//starte door-timer
					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println("---- Started doortimer")
					State = DOOROPEN

					//update elev-info
					//ElevState <- elevatorInfo
					updateFileAndElevator = true

				} else {
					//Restart motorTimer
					channels.StopImmobileTimer <- true
					go StoppableTimer(MAX_TRAVEL_TIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("---- Restarted Motortimer")
				}
			case DOOROPEN:
			case IMMOBILE:
				//Kommer den hit, da har den ikke motorstopp lenger, men den har heller ikke noe i køen
				//Start for motor!!

				
				if nextFloor == floorArrival{
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevatorInfo.Direction = elevio.MD_Stop

					//send a completed order message to OrderDistributed
					expidizeOrder(elevatorInfo, OrderUpdate)

					elevio.SetDoorOpenLamp(true)					
					removeFromQueue(&elevatorInfo)

					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println("---- Started doortimer")
					State = DOOROPEN

				} else {

					go StoppableTimer(MAX_TRAVEL_TIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("---- Started motortimer")
					State = MOVING

				}

				
				//si at jeg ikke har IMMOBILE lenger til Jon, da sender jon mine cabbies
				elevatorInfo.Immobile = false
				
				//ElevState <- elevatorInfo
				updateFileAndElevator = true

			}


		
		case <- channels.DoorTimedOut:
			fmt.Println("---- Door has timedout")
			fmt.Println("---- my state is : ", State)			
			
			switch State{
			case IDLE:
			case MOVING:
			case DOOROPEN:
				
				if obstructed == true{
					fmt.Println("---- OBSTRUCTION")

					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println("---- Restarted doortimer")

					if wasobstr == false {
					//starte obstruction timer  første gangen den er obsruert
					fmt.Println("---- started obstruction/immobility timer")
					go StoppableTimer(MAX_OBSTRUCTION_TIME, 1, channels.StopImmobileTimer, channels.Immobile)
					wasobstr = true 
					}
				}				
				if obstructed == false && wasobstr == true{
					fmt.Println("---- OBSTRUCTION OFF")
					fmt.Println("---- Stopping immobility timer 1")
					elevio.SetDoorOpenLamp(false) 
					channels.StopImmobileTimer <- true
					}

				if checkOrdersPresent(elevatorInfo) == true && obstructed == false{
					elevio.SetDoorOpenLamp(false)
					nextFloor = queueSearch(QueueDirection, elevatorInfo)

					dir := getDirection(elevatorInfo.CurrentFloor, nextFloor)
					elevio.SetMotorDirection(dir)
					QueueDirection = dir
					elevatorInfo.Direction = dir

					//start motor-timer
					go StoppableTimer(MAX_TRAVEL_TIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("---- Started motortimer")
					State = MOVING
				} else {
					if (obstructed == false && checkOrdersPresent(elevatorInfo) == false){
						elevio.SetDoorOpenLamp(false)
						State = IDLE
						fmt.Println("---- Im IDLE, have Closed door and NO MORE IN QUEUE")
					}
				}
				//ElevState <- elevatorInfo //update elev-info
				updateFileAndElevator = true


			case IMMOBILE:
				if obstructed == false {
					State = DOOROPEN
					elevatorInfo.Immobile = false
					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println("---- Restarted doortimer")
				}
				//ElevState <- elevatorInfo 
				updateFileAndElevator = true

			}


		case obstructed = <- channels.Obstruction:
				fmt.Println("---- Obstruction is : ", obstructed)
				if obstructed == false {
					wasobstr = false
					if State == DOOROPEN{
						channels.StopImmobileTimer <- true
					}
					if State == IMMOBILE{
						State = DOOROPEN
						resetDoor <- DOOR_OPEN_TIMER
						fmt.Println("---- Restarted doortimer, no longer obstructed")
					}
				}

		case <- channels.Immobile:
			fmt.Println( "---- IMMOBILITY detected")

			channels.StopImmobileTimer <- true
			elevatorInfo.Immobile = true
			emptyQueue(&elevatorInfo)
			immobilityNextFloor = nextFloor
			State = IMMOBILE
			//ElevState <- elevatorInfo
			updateFileAndElevator = true


		default:
		}
		if updateFileAndElevator == true{
			updateFileAndElevator = false

			writeToBackUpFile("CabOrders", ElevatorId, elevatorInfo)
			//writetoFile("cabOrders", LocalID, elevator)
			go func() { ElevState <- elevatorInfo }()

		}

		

	}

}



