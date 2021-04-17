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

func OrderDistributor(orderOut chan<- Order, orderIn chan Order, getElevatorState <-chan Elevator) {
	fmt.Println("*** Starting OrderDistributor...")
	var queue [NumberOfFloors]Order
	go pollOrders(orderIn)
	orderToNetworkChannel := make(chan Order)

	go orderNetworkCommunication(orderToNetworkChannel, orderIn)

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
			fmt.Println("*** reading order...")
			if queue[order.Floor].Status == NoActiveOrder && order.TimedOut {
				fmt.Println("*** expired order invalid: \t", order.Floor)
				break
			}

			switch order.Status {
			case NoActiveOrder:
			case WaitingForCost:
				fmt.Println("*** STATUS waiting for cost: \t", order.Floor)
 
				/* Logikk for å skru på lys i gangen hvis den har caborder der selv?
				if queue[order.Floor].CabOrder == true {
					if order.DirectionUp {
						elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, order.DirectionUp)
					}
					if order.DirectionDown {
						elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, order.DirectionDown)
					}
					go orderBuffer(order, orderToNetworkChannel)
					break
				} */

				if order.CabOrder == true {
					order.Cost[ElevatorId] = Costfunction(elevatorState, order) // Bruk costfunction
					//order.Status = Confirmed
					//order.CabOrder = false
					//orderToNetworkChannel <- order
					//order.CabOrder = true
					elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, true)
					//order.Status = Mine
					//queue[order.Floor] = order
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
				queue[order.Floor].Status = order.Status
				if queue[order.Floor].DirectionUp == false {
					queue[order.Floor].DirectionUp = order.DirectionUp
				}
				if queue[order.Floor].DirectionDown == false {
					queue[order.Floor].DirectionDown = order.DirectionDown
				}

				for elevator := 0; elevator < NumberOfElevators; elevator++ {
					if order.Cost[elevator] != MaxCost {
						queue[order.Floor].Cost[elevator] = order.Cost[elevator] // Sjekke om det ikke oppstår uenigheter
					}
				}
				if queue[order.Floor].Cost[ElevatorId] == MaxCost {
					fmt.Println("*** adding own cost: \t", order.Floor)
					queue[order.Floor].Cost[ElevatorId] = Costfunction(elevatorState, order)
					order = queue[order.Floor]
					order.CabOrder = false
					go orderBuffer(order, orderToNetworkChannel)
					fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
					go orderTimer(order, orderIn, 2)
					fmt.Println("-----------------------------------")
				}

				allCostsPresent := true
				fmt.Println("*** costcheck: \t", order.Floor)
				for elevator := 0; elevator < NumberOfElevators; elevator++ {
					fmt.Println(queue[order.Floor].Cost[elevator])
					if queue[order.Floor].Cost[elevator] == MaxCost {
						allCostsPresent = false
					}
				}

				if allCostsPresent || order.TimedOut {
					//fmt.Println("*** order with status WFC advances: \t", order.Floor)
					queue[order.Floor].Status = Unconfirmed
					go orderBuffer(queue[order.Floor], orderIn)
					go orderBuffer(queue[order.Floor], orderToNetworkChannel)
				}
				break

			case Unconfirmed:
				fmt.Println("*** STATUS Unconfirmed: \t", order.Floor)
				if queue[order.Floor].Status > Unconfirmed {
					fmt.Println("*** at higher status: \t", order.Floor)
					break
				}
				if order.TimedOut == true {
					queue[order.Floor].Cost[orderFindIdWithLowestCost(order)] = MaxCost
					go orderBuffer(queue[order.Floor], orderIn)
				}

				if orderFindIdWithLowestCost(order) == ElevatorId {
					fmt.Println("*** has LOWESTCOST: \t", order.Floor)
					queue[order.Floor].Status = Confirmed
					go orderBuffer(queue[order.Floor], orderToNetworkChannel)
					queue[order.Floor].Status = Mine
					go orderBuffer(queue[order.Floor], orderIn)
				} else {
					go orderTimer(queue[order.Floor], orderIn, queue[order.Floor].Cost[orderFindIdWithLowestCost(order)]*3+5)
				}
				break

			case Confirmed:
				// Sette på lys
				if order.DirectionUp == true {
					elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, true)
				}
				if order.DirectionDown == true {
					elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, true)
				}

				fmt.Println("*** STATUS Confirmed: \t", order.Floor)
				if queue[order.Floor].Status > Confirmed {
					fmt.Println("*** at higher status: \t", order.Floor)
					break
				}

				if order.TimedOut == true && queue[order.Floor].Status == Confirmed {
					if order.CabOrder == true {
						break // Må utbedres
					}
					queue[order.Floor].Status = Unconfirmed
					order.Status = Unconfirmed
					go orderBuffer(order, orderIn)
					break
				}

				queue[order.Floor] = order
				go orderTimer(order, orderIn, order.Cost[orderFindIdWithLowestCost(order)]*3+5) // Må endres til et uttrykk med costen
				break // Hva skjer hvis alle har MaxCost?

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
					go orderBuffer(order, orderToNetworkChannel)
					break
				} else if order.TimedOut == true {
					fmt.Println("*** ERROR: Could not expedite Caborder: \r", order.Floor)
					break
				}

				// send til fsm
				orderBuffer(queue[order.Floor], orderOut)
				fmt.Println("*** ORDER SENT TO FSM: \t", order.Floor)
				go orderTimer(order, orderIn, order.Cost[ElevatorId]*3+5) // Må også endres
				break

			case Done:
				fmt.Println("****** ORDER DONE: \t", order.Floor)
				if order.FromId == ElevatorId {
					go orderBuffer(order, orderToNetworkChannel)
					elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, false)
					order.CabOrder = false
				}
				elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, false)
				elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, false)

				order.Status = NoActiveOrder
				order.DirectionUp = false
				order.DirectionDown = false
				
				order.TimedOut = false
				for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
					order.Cost[elevatorNumber] = MaxCost
				}
				queue[order.Floor] = order
				break
			}

		case elevatorState = <- getElevatorState:
			if elevatorState.Immobile && !elevatorImmobile {
				fmt.Println("*** DO THING TO MAKE THE QUEUE BE BETTER SEÑOR!")
			}
			elevatorImmobile = elevatorState.Immobile
			break

		default:
		}
	}
}