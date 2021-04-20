package fsm

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	. "../timer"
	. "../types"
)

func InitFSM(numFloors int) {
	elevio.SetMotorDirection(elevio.MD_Down)
	for elevio.GetFloor() != 0 {
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetFloorIndicator(0)

	elevio.SetDoorOpenLamp(false)
	fmt.Println("FSM Initialized ")
}

func RunElevator(channels FsmChannels, OrderUpdate chan<- Order) {

	var elevatorInfo Elevator
	elevatorInfo.CurrentFloor = 0
	var nextFloor int
	var isObstructed bool
	var immobilityNextFloor int
	var queueDirection elevio.MotorDirection
	queueDirection = elevio.MD_Stop
	var obstructionCounter int
	obstructionCounter = MAX_DOOR_CLOSE_TRIES
	updateFileAndElevator := false

	resetDoor := make(chan time.Duration)
	go ResetableTimer(time.Duration(0), resetDoor, channels.DoorTimedOut)
	<-channels.DoorTimedOut
	readFromBackupFile("CabOrders", ELEVATOR_ID, &elevatorInfo)
	startupAfterCrash := checkIfOrdersPresentInQueue(elevatorInfo)
	
	State := IDLE

	for {
		select {
		case newOrder := <-channels.NewOrder:
			fmt.Println("----New order to floor: ", newOrder.Floor)

			switch State {
			case IDLE:
				if elevatorInfo.CurrentFloor == newOrder.Floor {
					elevio.SetDoorOpenLamp(true)
					resetDoor <- DOOR_OPEN_TIME_DURATION
					expediteOrder(elevatorInfo, OrderUpdate)
					State = DOOROPEN
				} else {
					addToQueue(&elevatorInfo, newOrder)
					goToNextInQueue(channels, elevatorInfo, &queueDirection, &nextFloor)
					State = MOVING
					updateFileAndElevator = true
				}
			case MOVING:
				addToQueue(&elevatorInfo, newOrder)

				if getNextFloorInQueue(queueDirection, elevatorInfo) == elevatorInfo.CurrentFloor {
					break
				} else {
					nextFloor = getNextFloorInQueue(queueDirection, elevatorInfo)
					fmt.Println("----my next floor is:", nextFloor)
				}
				updateFileAndElevator = true

			case DOOROPEN:
				if elevatorInfo.CurrentFloor == newOrder.Floor {
					elevio.SetDoorOpenLamp(true)
					resetDoor <- DOOR_OPEN_TIME_DURATION
					fmt.Println(" ---- Started Doortimer")
					removeFromQueue(&elevatorInfo)
					expediteOrder(elevatorInfo, OrderUpdate)

				} else {
					addToQueue(&elevatorInfo, newOrder)
					updateFileAndElevator = true

				}
			case IMMOBILE:
				if newOrder.CabOrder == true{
					fmt.Println("hei")
					readFromBackupFile("CabOrders", ELEVATOR_ID, &elevatorInfo)
					addToQueue(&elevatorInfo, newOrder) 
					writeToBackUpFile("CabOrders", ELEVATOR_ID, elevatorInfo)
				}
			}

		case floorArrival := <-channels.FloorReached:
			fmt.Println("---- Arriving at floor: ", floorArrival)
			elevatorInfo.CurrentFloor = floorArrival
			elevio.SetFloorIndicator(floorArrival)

			switch State {
			case IDLE:
			case MOVING:
				nextFloor = getNextFloorInQueue(queueDirection, elevatorInfo)
				if nextFloor == -1 {
					nextFloor = immobilityNextFloor
					fmt.Println("Using immobility nextFloor which is:", nextFloor)
				}

				if nextFloor == floorArrival {

					elevio.SetMotorDirection(elevio.MD_Stop)
					elevatorInfo.Direction = elevio.MD_Stop

					channels.StopImmobileTimer <- true
					fmt.Println("---- Stopped Motortimer")

					elevio.SetDoorOpenLamp(true)
					removeFromQueue(&elevatorInfo)
					expediteOrder(elevatorInfo, OrderUpdate)

					resetDoor <- DOOR_OPEN_TIME_DURATION
					fmt.Println("---- Started doortimer")
					State = DOOROPEN

					updateFileAndElevator = true

				} else {
					channels.StopImmobileTimer <- true
					go StoppableTimer(MAX_TRAVEL_TIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("---- Restarted Motortimer")
				}
			case DOOROPEN:
			case IMMOBILE:

				elevatorInfo.Immobile = false
				readFromBackupFile("CabOrders", ELEVATOR_ID, &elevatorInfo)

				if immobilityNextFloor == floorArrival {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevatorInfo.Direction = elevio.MD_Stop

					expediteOrder(elevatorInfo, OrderUpdate)

					elevio.SetDoorOpenLamp(true)
					removeFromQueue(&elevatorInfo)

					resetDoor <- DOOR_OPEN_TIME_DURATION
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
			switch State {
			case IDLE:
			case MOVING:
			case DOOROPEN:
				expediteOrder(elevatorInfo, OrderUpdate)

				if isObstructed == true {
					fmt.Println("---- OBSTRUCTION")
					obstructionCounter--
					fmt.Println("---- ", obstructionCounter, " tries left!")
					resetDoor <- DOOR_OPEN_TIME_DURATION
				} else {
					obstructionCounter = MAX_DOOR_CLOSE_TRIES
				}
				if obstructionCounter == 0 {
					State = IMMOBILE
					elevatorInfo.Immobile = true
					obstructionCounter = MAX_DOOR_CLOSE_TRIES
				}

				if checkIfOrdersPresentInQueue(elevatorInfo) == true && isObstructed == false {
					elevio.SetDoorOpenLamp(false)
					goToNextInQueue(channels, elevatorInfo, &queueDirection, &nextFloor)
					State = MOVING
				} else {
					if isObstructed == false && checkIfOrdersPresentInQueue(elevatorInfo) == false {
						elevio.SetDoorOpenLamp(false)
						State = IDLE
						fmt.Println("---- Im IDLE, have Closed door and NO MORE IN QUEUE")
					}
				}
				updateFileAndElevator = true

			case IMMOBILE:
				if isObstructed == false {
					State = DOOROPEN
					elevatorInfo.Immobile = false
					readFromBackupFile("CabOrders", ELEVATOR_ID, &elevatorInfo)
					resetDoor <- DOOR_OPEN_TIME_DURATION
					fmt.Println("---- Restarted doortimer")
				}
				updateFileAndElevator = true
			}

		case isObstructed = <-channels.Obstruction:
			fmt.Println("---- Obstruction is : ", isObstructed)
			if isObstructed == false && State == IMMOBILE {
				State = DOOROPEN
				elevatorInfo.Immobile = false
				resetDoor <- DOOR_OPEN_TIME_DURATION
				fmt.Println("---- Restarted doortimer, no longer isObstructed")
				updateFileAndElevator = true
			}

		case <-channels.Immobile:
			fmt.Println("---- IMMOBILITY detected")

			channels.StopImmobileTimer <- true
			elevatorInfo.Immobile = true
			State = IMMOBILE
			updateFileAndElevator = true
			immobilityNextFloor = nextFloor

		default:
			if startupAfterCrash == true {
				startupAfterCrash = false
				goToNextInQueue(channels, elevatorInfo, &queueDirection, &nextFloor)
				State = MOVING
			}
		}
		if updateFileAndElevator == true {
			updateFileAndElevator = false

			writeToBackUpFile("CabOrders", ELEVATOR_ID, elevatorInfo)
			go func() { channels.ElevatorState <- elevatorInfo }()
		}
	}
}
