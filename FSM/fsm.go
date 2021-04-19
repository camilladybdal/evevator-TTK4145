package fsm

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	. "../timer"
	. "../types"
)

func RunElevator(channels FsmChannels, OrderUpdate chan<- Order, ElevState chan<- Elevator) {
	State := IDLE
	var elevatorInfo Elevator
	emptyQueue(&elevatorInfo)

	elevatorInfo.CurrentFloor = 0
	wasobstr := false
	updateFileAndElevator := false
	startedObstructionTimer := false


	var QueueDirection elevio.MotorDirection
	QueueDirection = elevio.MD_Stop

	//newCountdownTime <-chan time.Duration
	resetDoor := make(chan time.Duration)
	go ResetableTimer(time.Duration(0), resetDoor, channels.DoorTimedOut)
	<-channels.DoorTimedOut

	var nextFloor int
	var obstructed bool
	var immobilityNextFloor int

	//read cab-orders from file and add to queues.
	readFromBackupFile("CabOrders", ElevatorId, &elevatorInfo)
	startupAfterCrash := checkOrdersPresent(elevatorInfo)

	go elevio.PollFloorSensor(channels.FloorReached)
	go elevio.PollObstructionSwitch(channels.Obstruction)
	fmt.Println("Polling started...")

	//for select switch case
	for {
		select {
		case newOrder := <-channels.NewOrder:
			fmt.Println("----New order to floor: ", newOrder.Floor)
			fmt.Println("---- my State is: ", State)

			switch State {
			case IDLE:
				//sjekk om du er i den etasjen fra før av
				if elevatorInfo.CurrentFloor == newOrder.Floor {

					elevio.SetDoorOpenLamp(true)
					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println("im here 1")
					fmt.Println("---- Started Doortimer")

					expidizeOrder(elevatorInfo, OrderUpdate)
					
					
					/*FOR TESTING PURPOSES*/

					State = DOOROPEN

				} else {

					//legger til i køen
					addToQueue(&elevatorInfo, newOrder)
					goToNextInQueue(channels, elevatorInfo, &QueueDirection, &nextFloor)
					State = MOVING

					//update elev-info
					updateFileAndElevator = true
				}
			case MOVING:
				//legger til i køen
				addToQueue(&elevatorInfo, newOrder)

				if queueSearch(QueueDirection, elevatorInfo) == elevatorInfo.CurrentFloor {
					break
				} else {
					//her sjekker den jo ikke om køen er tom! fordi den er moving. Så kanskje den er moving på feil sted?
					nextFloor = queueSearch(QueueDirection, elevatorInfo)
					fmt.Println("----my next floor is:", nextFloor)
				}

				//update elev-info
				updateFileAndElevator = true

			case DOOROPEN:
				if elevatorInfo.CurrentFloor == newOrder.Floor {

					elevio.SetDoorOpenLamp(true)

					//Reset this timer, dont start a new
					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println(" ---- Started Doortimer")


					fmt.Println("im here 2")
					removeFromQueue(&elevatorInfo)

					expidizeOrder(elevatorInfo, OrderUpdate)

				} else {
					addToQueue(&elevatorInfo, newOrder)
					updateFileAndElevator = true

				}
			case IMMOBILE:
			}

		case floorArrival := <-channels.FloorReached:
			fmt.Println("---- Arriving at floor: ", floorArrival)

			elevatorInfo.CurrentFloor = floorArrival
			elevio.SetFloorIndicator(floorArrival)

			switch State {
			case IDLE:
			case MOVING:

				nextFloor = queueSearch(QueueDirection, elevatorInfo)

				if nextFloor == -1 {
					nextFloor = immobilityNextFloor
					fmt.Println("Using immobility nextFloor which is:", nextFloor)
				}

				fmt.Println("---- I am heding for this floor: ", nextFloor)

				if nextFloor == floorArrival {

					elevio.SetMotorDirection(elevio.MD_Stop)
					elevatorInfo.Direction = elevio.MD_Stop

					//Stop motorTimer
					channels.StopImmobileTimer <- true
					fmt.Println("---- Stopped Motortimer")

					elevio.SetDoorOpenLamp(true)
					removeFromQueue(&elevatorInfo)
					expidizeOrder(elevatorInfo, OrderUpdate)

					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println("---- Started doortimer")
					State = DOOROPEN

					updateFileAndElevator = true

				} else {
					//Restart motorTimer
					channels.StopImmobileTimer <- true
					go StoppableTimer(MAX_TRAVEL_TIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("---- Restarted Motortimer")
				}
			case DOOROPEN:
			case IMMOBILE:

				elevatorInfo.Immobile = false
				readFromBackupFile("CabOrders", ElevatorId, &elevatorInfo)

				if immobilityNextFloor == floorArrival {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevatorInfo.Direction = elevio.MD_Stop

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
				updateFileAndElevator = true
			}

		case <-channels.DoorTimedOut:
			fmt.Println("---- Door has timedout")
			fmt.Println("---- my state is : ", State)

			switch State {
			case IDLE:
			case MOVING:
			case DOOROPEN:

				expidizeOrder(elevatorInfo, OrderUpdate)

				if obstructed == true {
					fmt.Println("---- OBSTRUCTION")

					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println("---- Restarted doortimer")

					fmt.Println("wasobstr is : ", wasobstr)

					if wasobstr == false {
						fmt.Println("---- started obstruction/immobility timer")
						go StoppableTimer(MAX_OBSTRUCTION_TIME, 1, channels.StopImmobileTimer, channels.Immobile)
						wasobstr = true
						startedObstructionTimer = true
					}
				}
				if obstructed == false && wasobstr == true {
					fmt.Println("---- OBSTRUCTION OFF")
					fmt.Println("---- Stopping immobility timer 1")
					elevio.SetDoorOpenLamp(false)
					channels.StopImmobileTimer <- true
				}

				if checkOrdersPresent(elevatorInfo) == true && obstructed == false {
					elevio.SetDoorOpenLamp(false)
					goToNextInQueue(channels, elevatorInfo, &QueueDirection, &nextFloor)
					State = MOVING
				} else {
					if obstructed == false && checkOrdersPresent(elevatorInfo) == false {
						elevio.SetDoorOpenLamp(false)
						State = IDLE
						fmt.Println("---- Im IDLE, have Closed door and NO MORE IN QUEUE")
					}
				}
				updateFileAndElevator = true

			case IMMOBILE:
				if obstructed == false {
					State = DOOROPEN
					elevatorInfo.Immobile = false
					readFromBackupFile("CabOrders", ElevatorId, &elevatorInfo)
					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println("---- Restarted doortimer")
				}
				updateFileAndElevator = true
			}

		case obstructed = <-channels.Obstruction:
			fmt.Println("---- Obstruction is : ", obstructed)
			
			if obstructed == false {
				wasobstr = false
				if State == DOOROPEN && startedObstructionTimer == true{
					channels.StopImmobileTimer <- true
					startedObstructionTimer = false
				}
				if State == IMMOBILE {
					State = DOOROPEN
					elevatorInfo.Immobile = false
					resetDoor <- DOOR_OPEN_TIMER
					fmt.Println("---- Restarted doortimer, no longer obstructed")
				}
				updateFileAndElevator = false
			}

		case <-channels.Immobile:
			fmt.Println("---- IMMOBILITY detected")

			channels.StopImmobileTimer <- true
			elevatorInfo.Immobile = true
			State = IMMOBILE
			updateFileAndElevator = true
			startedObstructionTimer = false
			immobilityNextFloor = nextFloor
			fmt.Println("immobilitynextfloor is: ", immobilityNextFloor)



		default:
			if startupAfterCrash == true {
				startupAfterCrash = false
				goToNextInQueue(channels, elevatorInfo, &QueueDirection, &nextFloor)
				State = MOVING
			}
		}
		if updateFileAndElevator == true {
			updateFileAndElevator = false

			writeToBackUpFile("CabOrders", ElevatorId, elevatorInfo)
			go func() { ElevState <- elevatorInfo }()
		}
	}
}
