package fsm

import (
	"../elevio"
	"fmt"
	. "../types"
	. "../timer"
	. "../config"
	//"time"
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

	fmt.Println("FSM Initialized ")
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
			fmt.Println("----New order to floor: ", newOrder.Floor)
			fmt.Println("---- my State is: ", State)

			switch State{
			case IDLE:
				//sjekk om du er i den etasjen fra før av
				if elevatorInfo.CurrentFloor == newOrder.Floor{

					elevio.SetDoorOpenLamp(true)
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut) 
					fmt.Println("---- Started Doortimer")

					newOrder.Status = Done 
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

					/*TESTING IF QUEUE-ALGORITHM IS CORRECT*/

					/*

					fmt.Println("---- My queue after adding the new order to it is: ")

					fmt.Println("Upqueue:: ")
					for i:=0;i<NumFloors;i++{
						fmt.Println(elevatorInfo.UpQueue[i])
					}
					fmt.Println("Downqueue: ")
					for i:=0;i<NumFloors;i++{
					fmt.Println(elevatorInfo.DownQueue[i])
					}
					*/

					//Søker etter floor HVIS den er i IDLE
					nextFloor = queueSearch(QueueDirection, elevatorInfo)
					fmt.Println("---- floor im heading for is: ", nextFloor)

					
					/*-----------------------------------------------*/


					dir := getDirection(elevatorInfo.CurrentFloor, nextFloor)
					elevio.SetMotorDirection(dir)
					
					QueueDirection = dir
					elevatorInfo.Direction = dir	
					fmt.Println("---- direction to floor is: " , dir)

					
					//Start for motor!!
					go StoppableTimer(PASSINGFLOORTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("---- Started motortimer")
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

				if queueSearch(QueueDirection, elevatorInfo) == elevatorInfo.CurrentFloor{
					break;
				} else{
					nextFloor = queueSearch(QueueDirection, elevatorInfo)
					fmt.Println("----my next floor is:", nextFloor)
				}

				//update elev-info
				ElevState <- elevatorInfo	
					
			case DOOROPEN:
				if elevatorInfo.CurrentFloor == newOrder.Floor{

					elevio.SetDoorOpenLamp(true)
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut) 
					fmt.Println(" ---- Started Doortimer")
					removeFromQueue(&elevatorInfo)

					//send a completed order message to OrderDistributed
					newOrder.Status = Done 
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
			fmt.Println("---- Arriving at floor: ", floorArrival)

			elevatorInfo.CurrentFloor = floorArrival
			elevio.SetFloorIndicator(floorArrival)

			switch State{
			case IDLE:
			case MOVING:

				nextFloor = queueSearch(QueueDirection, elevatorInfo)
				fmt.Println("---- I am heding for this floor: ", nextFloor)

				/*
				fmt.Println("---- My queue after adding the new order to it is: ")

					fmt.Println("Upqueue:: ")
					for i:=0;i<NumFloors;i++{
						fmt.Println(elevatorInfo.UpQueue[i])
					}
					fmt.Println("Downqueue: ")
					for i:=0;i<NumFloors;i++{
						fmt.Println(elevatorInfo.DownQueue[i])
					}

				*/

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
					var Expidized_order Order
					Expidized_order.Floor = floorArrival
					Expidized_order.Status = Done //replace with Done
					Expidized_order.FromId = ElevatorId
					OrderUpdate <- Expidized_order

					//starte door-timer
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.Println("---- Started doortimer")
					State = DOOROPEN

					//update elev-info
					ElevState <- elevatorInfo
				} else {
					//Restart motorTimer
					channels.StopImmobileTimer <- true
					go StoppableTimer(PASSINGFLOORTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("---- Restarted Motortimer")
				}
			case DOOROPEN:
			case IMMOBILE:
				//Kommer den hit, da har den ikke motorstopp lenger, men den har heller ikke noe i køen
				//Start for motor!!

				
				if nextFloor == floorArrival{
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevatorInfo.Direction = elevio.MD_Stop

					//skal denne stå her?
					var Expidized_order Order
					Expidized_order.Floor = floorArrival
					Expidized_order.Status = Done 
					Expidized_order.FromId = ElevatorId
					OrderUpdate <- Expidized_order

					elevio.SetDoorOpenLamp(true)					
					removeFromQueue(&elevatorInfo)

					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.Println("---- Started doortimer")
					State = DOOROPEN

				} else {

					go StoppableTimer(PASSINGFLOORTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("---- Started motortimer")
					State = MOVING

				}

				
				//si at jeg ikke har IMMOBILE lenger til Jon, da sender jon mine cabbies
				elevatorInfo.Immobile = false
				ElevState <- elevatorInfo
			}


		
		case <- channels.DoorTimedOut:
			fmt.Println("---- Door has timedout")
			fmt.Println("---- my state is : ", State)			
			
			switch State{
			case IDLE:
			case MOVING:
			case DOOROPEN:
				
				/*ER DET HER PROBLEMET LIGGER?*/
				if obstructed == true{
					fmt.Println("---- OBSTRUCTION")

					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.Println("---- Restarted doortimer")

					if wasobstr == false {
					//starte obstruction timer  første gangen den er obsruert
					fmt.Println("---- started obstruction/immobility timer")
					go StoppableTimer(MAXOBSTRUCTIONTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					wasobstr = true 
					}

					//nå er denne forever true frem til hele driten omgangen er ferdig
				}

				
				//stop obstruction timer
				if obstructed == false && wasobstr == true{
					fmt.Println("---- OBSTRUCTION OFF")

					fmt.Println("---- Stopping immobility timer 1")
					elevio.SetDoorOpenLamp(false) 
					channels.StopImmobileTimer <- true
					}
				

				if checkOrdersPresent(elevatorInfo) == true && obstructed == false{
					elevio.SetDoorOpenLamp(false)
					nextFloor = queueSearch(QueueDirection, elevatorInfo)

					/*FOR TESTING PURPOSES*/
					for i:=0;i<NumFloors;i++{
						fmt.Println(elevatorInfo.UpQueue[i])
					}
					for i:=0;i<NumFloors;i++{
					fmt.Println(elevatorInfo.DownQueue[i])
					}


					dir := getDirection(elevatorInfo.CurrentFloor, nextFloor)
					elevio.SetMotorDirection(dir)
					QueueDirection = dir
					elevatorInfo.Direction = dir

					//start motor-timer
					go StoppableTimer(PASSINGFLOORTIME, 1, channels.StopImmobileTimer, channels.Immobile)
					fmt.Println("---- Started motortimer")
					State = MOVING


				} else {
					if (obstructed == false && checkOrdersPresent(elevatorInfo) == false){
						elevio.SetDoorOpenLamp(false)
						State = IDLE
						fmt.Println("---- Im IDLE, have Closed door and NO MORE IN QUEUE")
					}
				}

				//update elev-info
				ElevState <- elevatorInfo

			case IMMOBILE:
				

				if obstructed == false {
					State = DOOROPEN
					elevatorInfo.Immobile = false
					go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
					fmt.Println("---- Restarted doortimer")

				}
				ElevState <- elevatorInfo 
				//elevstaten har jo ikke endret seg nå...
			}


		case obstructed = <- channels.Obstruction:
				fmt.Println("---- Obstruction is : ", obstructed)

				if obstructed == false {
					wasobstr = false
					if State == DOOROPEN{
						channels.StopImmobileTimer <- true
					}

					/*Dette er hvis den er immobil, og dør-timeren ikke går lenger*/
					if State == IMMOBILE{
						State = DOOROPEN
						go CountDownTimer(DOOROPENTIME, channels.DoorTimedOut)
						fmt.Println("---- Restarted doortimer, no longer obstructed")
					}
				}

		case <- channels.Immobile:
			fmt.Println( "---- IMMOBILITY detected")

			//stop immobility timer
			channels.StopImmobileTimer <- true

			//tell OrderDistributer that I am IMMOBILE 
			elevatorInfo.Immobile = true

			//tømme min egen kø, jon sender den til andre heiser
			emptyQueue(&elevatorInfo)

			fmt.Println("---- Queue should be empty now: ")
			fmt.Println("----nextfloor is:" ,nextFloor)

		

			for i:=0;i<NumFloors;i++{
				fmt.Println(elevatorInfo.UpQueue[i])
			}
			for i:=0;i<NumFloors;i++{
			fmt.Println(elevatorInfo.DownQueue[i])
			}

			//update elevInfo
			State = IMMOBILE
			ElevState <- elevatorInfo

		
		default:
		}
	}
}



