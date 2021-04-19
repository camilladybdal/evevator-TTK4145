package orderDistributor

import (
	"fmt"
	"time"
	"../timer"
	. "../config"
	"../elevio"
	//"../network/bcast"
	. "../types"
	. "../costfnc"
)



//func orderDumpQueue(queue *[]Order) {}


/*
func orderDumpQueue(queue *[]Order) {
	for floor := 0; floor < NumberOfFloors; floor++ {

	}
}
*/

func OrderDistributor(orderOut chan<- Order, orderIn chan Order, getElevatorState <-chan Elevator, newButtonEvent chan elevio.ButtonEvent, networkTransmit chan Order, networkRecieve <-chan Order) {
	fmt.Println("*** Starting OrderDistributor...")
	var queue [NumberOfFloors]Order
	
	orderToNetwork := make(chan Order)
	go pollOrders(orderIn, newButtonEvent)
	go orderNetworkCommunication(networkTransmit, networkRecieve, orderToNetwork, orderIn)

	var elevatorState Elevator
	var elevatorImmobile bool

	for floor := 0; floor < NumberOfFloors; floor++ {
		queue[floor].Floor = floor
		queue[floor].DirectionUp = false
		queue[floor].DirectionDown = false
		queue[floor].CabOrder = false
		for elevator := 0; elevator < NumberOfElevators; elevator++ {
			queue[floor].Cost[elevator] = MaxCost
		}
		queue[floor].Status = NoActiveOrder
		queue[floor].TimedOut = false
		queue[floor].FromId = ElevatorId
		queue[floor].Timestamp = time.Now().Unix()

		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	}

	// For handling deadlock in orderIn
	startDraining := make(chan bool)
	refreshTimer := make(chan time.Duration)
	resetTime := time.Duration(10)
	go timer.ResetableTimer(resetTime, refreshTimer, startDraining)
	go drainChannels(orderIn, startDraining)

	for {
		refreshTimer <- resetTime
		
		select {
		// Order pipeline
		case order := <-orderIn:

			if queue[order.Floor].Status >= Confirmed && elevatorState.CurrentFloor != order.Floor {
				if queue[order.Floor].DirectionUp == false && order.DirectionUp == true {
					elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, true)
					queue[order.Floor].DirectionUp = true
					go orderBuffer(order, orderToNetwork)
				}
				if queue[order.Floor].DirectionDown == false && order.DirectionDown == true {
					elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, true)
					queue[order.Floor].DirectionDown = true
					go orderBuffer(order, orderToNetwork)
				}
			}
			if queue[order.Floor].Status == NoActiveOrder && order.TimedOut {
				//fmt.Println("*** expired order invalid: \t", order.Floor)
				break
			}
			if elevatorImmobile && order.CabOrder == false && order.Status != Done && order.Status != WaitingForCost {
				fmt.Println("*** Elevator is immobile")
				break
			}
			if order.TimedOut && order.Status != queue[order.Floor].Status {
				break
			}
			if order.Timestamp < queue[order.Floor].Timestamp && order.TimedOut && order.FromId != ElevatorId {
				break
			}

			switch order.Status {
			case NoActiveOrder:
			case WaitingForCost:
				fmt.Println("*** STATUS waiting for cost: \t", order.Floor)
 
				if order.CabOrder == true {
					elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, true)
					if queue[order.Floor].CabOrder == false {
						queue[order.Floor].CabOrder = true
						fmt.Println("*** sent caborder to FSM")
						go orderBuffer(order, orderOut)
					}
					break
				}

				if queue[order.Floor].Status > WaitingForCost {
					fmt.Println("*** at higher status: \t", order.Floor, queue[order.Floor].Status)
					break
				}

				if queue[order.Floor].DirectionUp == false && queue[order.Floor].DirectionDown == false {
					queue[order.Floor].DirectionUp = order.DirectionUp
					queue[order.Floor].DirectionDown = order.DirectionDown
				}

				for elevator := 0; elevator < NumberOfElevators; elevator++ {
					if order.Cost[elevator] != MaxCost {
						queue[order.Floor].Cost[elevator] = order.Cost[elevator] // Sjekke om det ikke oppstår uenigheter
					}
				}
				if queue[order.Floor].Status != WaitingForCost {
					fmt.Println("*** adding own cost: \t", order.Floor)
					if elevatorState.Immobile != true {
						queue[order.Floor].Cost[ElevatorId] = Costfunction(elevatorState, order)
						order.Cost[ElevatorId] = queue[order.Floor].Cost[ElevatorId]
					} else {
						queue[order.Floor].Cost[ElevatorId] = MaxCost
						order.Cost[ElevatorId] = MaxCost
					}
					
					go orderBuffer(order, orderToNetwork)
					go orderTimer(order, orderIn, 1)
				}
				queue[order.Floor].Status = order.Status
				allCostsPresent := true
				fmt.Println("*** costcheck: \t", order.Floor)
				for elevator := 0; elevator < NumberOfElevators; elevator++ {
					fmt.Println(queue[order.Floor].Cost[elevator])
					if queue[order.Floor].Cost[elevator] == MaxCost {
						allCostsPresent = false
					}
				}

				if allCostsPresent || order.TimedOut {
					queue[order.Floor].Status = Unconfirmed
					go orderBuffer(queue[order.Floor], orderIn)
					//go orderBuffer(queue[order.Floor], orderToNetwork)
					// Linjen over er kun redundancy, legg til hvis det ser ut som om det kan være et problem
				}

			case Unconfirmed:
				fmt.Println("*** STATUS Unconfirmed: \t", order.Floor)
				if queue[order.Floor].Status > Unconfirmed {
					fmt.Println("*** at higher status: \t", order.Floor)
					break
				}
				if order.TimedOut == true {
					queue[order.Floor].Cost[orderFindIdWithLowestCost(order)] = MaxCost
					go orderBuffer(queue[order.Floor], orderIn)
					break
				}

				if orderFindIdWithLowestCost(order) == ElevatorId {
					fmt.Println("*** has LOWESTCOST: \t", order.Floor)
					if elevatorState.CurrentFloor != order.Floor || elevatorState.Direction == elevio.MD_Stop {
						queue[order.Floor].Status = Confirmed
						order.Status = Confirmed
						go orderBuffer(order, orderToNetwork)
					}
					queue[order.Floor].Status = Mine
					go orderBuffer(queue[order.Floor], orderIn)
				} else {
					go orderTimer(queue[order.Floor], orderIn, 5)
				}
				break

			case Confirmed:
				if queue[order.Floor].Status == NoActiveOrder {
					break
				}
				// Sette på lys
				if order.DirectionUp == true {
					elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, true)
				}
				if order.DirectionDown == true {
					elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, true)
				}
				if queue[order.Floor].Status < Confirmed && order.TimedOut == false {
					go orderBuffer(order, orderToNetwork)
				}
			
				fmt.Println("*** STATUS Confirmed: \t", order.Floor)
				if queue[order.Floor].Status > Confirmed {
					fmt.Println("*** at higher status: \t", order.Floor)
					break
				}

				if order.TimedOut == true && queue[order.Floor].Status == Confirmed {
					/*
					if order.CabOrder == true {
						break // Må utbedres
					}
					*/
					queue[order.Floor].Status = Unconfirmed
					order.Status = Unconfirmed
					go orderBuffer(order, orderIn)
					break
				}
				if queue[order.Floor].Status != Confirmed {
					queue[order.Floor].Status = Confirmed
					go orderTimer(queue[order.Floor], orderIn, order.Cost[orderFindIdWithLowestCost(order)]+DOOR_OPEN_TIME*NumberOfFloors)
				}
				//queue[order.Floor].Status = order.Status
				 // Må endres til et uttrykk med costen
				// Hva skjer hvis alle har MaxCost?

			case Mine:
				fmt.Println("*** STATUS Mine: \t", order.Floor)
				if queue[order.Floor].Status > Mine {
					fmt.Println("*** order with status Mine CANCELLED: \t", order.Floor)
					break
				}

				if order.TimedOut == true && order.CabOrder == false {
					fmt.Println("** order with status Mine TIMEDOUT: \t", order.Floor)
					order.Status = Confirmed
					queue[order.Floor].Status = Confirmed
					go orderBuffer(order, orderToNetwork)
					break
				} else if order.TimedOut == true {
					fmt.Println("*** ERROR: Could not expedite Caborder: \r", order.Floor)
					break
				}

				// send til fsm
				if order.DirectionUp == true {
					queue[order.Floor].DirectionUp = true
				}
				if order.DirectionDown == true {
					queue[order.Floor].DirectionDown = true
				}
				go orderBuffer(queue[order.Floor], orderOut)
				fmt.Println("*** ORDER SENT TO FSM: \t", order.Floor)
				go orderTimer(order, orderIn, order.Cost[ElevatorId]*3+5)
				break

			case Done:
				fmt.Println("****** ORDER DONE: \t", order.Floor)
				if order.FromId == ElevatorId {
					go orderBuffer(order, orderToNetwork)
					elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, false)
					order.CabOrder = false
				}
				elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, false)
				elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, false)

				order.Status = NoActiveOrder
				order.DirectionUp = false
				order.DirectionDown = false
				
				order.TimedOut = false
				order.Timestamp = time.Now().Unix()
				for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
					order.Cost[elevatorNumber] = MaxCost
				}
				queue[order.Floor] = order
				break
			}

		case elevatorState = <- getElevatorState:
			if elevatorState.Immobile && !elevatorImmobile {
				for floor := 0; floor < NumberOfFloors; floor++ {
					if queue[floor].Status == Mine {
						//queue[floor].Cost[ElevatorId] = MaxCost
						queue[floor].Status = Confirmed
						queue[floor].TimedOut = true
						go orderBuffer(queue[floor], orderToNetwork)
						queue[floor].DirectionUp = false
						queue[floor].DirectionDown = false
						
						// Må queue cleares også?
					}
				}
			}
			elevatorImmobile = elevatorState.Immobile
			break

		default:
		}
	}
}