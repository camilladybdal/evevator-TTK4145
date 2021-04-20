package orderDistributor

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	"../timer"

	. "../costfnc"
	. "../types"
)

func OrderDistributor(channels OrderDistributorChannels, orderOut chan<- Order, getElevatorState <-chan Elevator) {
	fmt.Println("*** Starting OrderDistributor...")
	var allOrders [NUMBER_OF_FLOORS]Order

	orderToPeers := make(chan Order)
	go pollOrders(channels.OrderUpdate, channels.NewButtonEvent)
	go orderCommunication(channels.OrderTransmitter, channels.OrderReciever, orderToPeers, channels.OrderUpdate)

	var elevatorState Elevator
	var elevatorIsImmobile bool

	for floor := 0; floor < NUMBER_OF_FLOORS; floor++ {
		allOrders[floor].Floor = floor
		allOrders[floor].DirectionUp = false
		allOrders[floor].DirectionDown = false
		allOrders[floor].CabOrder = false
		for elevator := 0; elevator < NUMBER_OF_ELEVATORS; elevator++ {
			allOrders[floor].Cost[elevator] = MAXCOST
		}
		allOrders[floor].Status = NotActive
		allOrders[floor].TimedOut = false
		allOrders[floor].FromId = ELEVATOR_ID
		allOrders[floor].Timestamp = time.Now().Unix()

		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	}

	// For handling deadlock in channels.OrderUpdate
	startDraining := make(chan bool)
	refreshTimer := make(chan time.Duration)
	resetTime := time.Duration(10)
	go timer.ResetableTimer(resetTime, refreshTimer, startDraining)
	go drainChannels(channels.OrderUpdate, startDraining)

	for {
		refreshTimer <- resetTime

		select {
		// Order pipeline
		case order := <-channels.OrderUpdate:

			/* Turns on hall light if there already is an order in the opposite direction */
			if allOrders[order.Floor].Status >= Confirmed && elevatorState.CurrentFloor != order.Floor {
				if allOrders[order.Floor].DirectionUp == false && order.DirectionUp == true {
					elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, true)
					allOrders[order.Floor].DirectionUp = true
					go orderBuffer(order, orderToPeers)
				}
				if allOrders[order.Floor].DirectionDown == false && order.DirectionDown == true {
					elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, true)
					allOrders[order.Floor].DirectionDown = true
					go orderBuffer(order, orderToPeers)
				}
			}

			/* Disregards orders that time out after completion, or are otherwise not wanted to treat */
			if allOrders[order.Floor].Status == NotActive && order.TimedOut {
				break
			}
			if elevatorIsImmobile && order.CabOrder == false && order.Status != Done && order.Status != WaitingForCost {
				break
			}
			if order.TimedOut && order.Status != allOrders[order.Floor].Status {
				break
			}
			if order.Timestamp < allOrders[order.Floor].Timestamp && order.TimedOut && order.FromId != ELEVATOR_ID {
				break
			}

			switch order.Status {
			case NotActive:
			case WaitingForCost:
				fmt.Println("*** STATUS waiting for cost: \t", order.Floor)

				if order.CabOrder == true {
					elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, true)
					if allOrders[order.Floor].CabOrder == false {
						allOrders[order.Floor].CabOrder = true
						fmt.Println("*** sent caborder to FSM")
						go orderBuffer(order, orderOut)
					}
					break
				}

				if allOrders[order.Floor].Status > WaitingForCost {
					fmt.Println("*** at higher status: \t", order.Floor, allOrders[order.Floor].Status)
					break
				}
				// ???
				if allOrders[order.Floor].DirectionUp == false && allOrders[order.Floor].DirectionDown == false {
					allOrders[order.Floor].DirectionUp = order.DirectionUp
					allOrders[order.Floor].DirectionDown = order.DirectionDown
				}

				for elevator := 0; elevator < NUMBER_OF_ELEVATORS; elevator++ {
					if order.Cost[elevator] != MAXCOST {
						allOrders[order.Floor].Cost[elevator] = order.Cost[elevator]
					}
				}

				if allOrders[order.Floor].Status != WaitingForCost {
					fmt.Println("*** adding own cost: \t", order.Floor)
					if elevatorState.Immobile != true {
						allOrders[order.Floor].Cost[ELEVATOR_ID] = Costfunction(elevatorState, order)
						order.Cost[ELEVATOR_ID] = allOrders[order.Floor].Cost[ELEVATOR_ID]
					} else {
						allOrders[order.Floor].Cost[ELEVATOR_ID] = MAXCOST
						order.Cost[ELEVATOR_ID] = MAXCOST
					}

					go orderBuffer(order, orderToPeers)
					go timingOrder(order, channels.OrderUpdate, 1)
				}

				allOrders[order.Floor].Status = order.Status
				allCostsPresent := true
				fmt.Println("*** costcheck: \t", order.Floor)
				for elevator := 0; elevator < NUMBER_OF_ELEVATORS; elevator++ {
					fmt.Println(allOrders[order.Floor].Cost[elevator])
					if allOrders[order.Floor].Cost[elevator] == MAXCOST {
						allCostsPresent = false
					}
				}

				if allCostsPresent || order.TimedOut {
					allOrders[order.Floor].Status = Unconfirmed
					go orderBuffer(allOrders[order.Floor], channels.OrderUpdate)
				}

			case Unconfirmed:
				fmt.Println("*** STATUS Unconfirmed: \t", order.Floor)
				if allOrders[order.Floor].Status > Unconfirmed {
					fmt.Println("*** at higher status: \t", order.Floor)
					break
				}
				if order.TimedOut == true {
					allOrders[order.Floor].Cost[findIdWithLowestCost(order)] = MAXCOST
					go orderBuffer(allOrders[order.Floor], channels.OrderUpdate)
					break
				}

				if findIdWithLowestCost(order) == ELEVATOR_ID {
					fmt.Println("*** I have LOWEST COST: \t", order.Floor)
					if elevatorState.CurrentFloor != order.Floor || elevatorState.Direction == elevio.MD_Stop {
						allOrders[order.Floor].Status = Confirmed
						order.Status = Confirmed
						go orderBuffer(order, orderToPeers)
					}
					allOrders[order.Floor].Status = Mine
					go orderBuffer(allOrders[order.Floor], channels.OrderUpdate)
				} else {
					go timingOrder(allOrders[order.Floor], channels.OrderUpdate, 5)
				}

			case Confirmed:
				/* Without this statement to break we might get a delayed answer and end up taking an order too many times,
				   drastically improves performance */
				if allOrders[order.Floor].Status == NotActive {
					break
				}

				if order.DirectionUp == true {
					elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, true)
				}
				if order.DirectionDown == true {
					elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, true)
				}
				if allOrders[order.Floor].Status < Confirmed && order.TimedOut == false {
					go orderBuffer(order, orderToPeers)
				}

				fmt.Println("*** STATUS Confirmed: \t", order.Floor)
				if allOrders[order.Floor].Status > Confirmed {
					fmt.Println("*** at higher status: \t", order.Floor)
					break
				}

				if order.TimedOut == true && allOrders[order.Floor].Status == Confirmed {
					allOrders[order.Floor].Status = Unconfirmed
					order.Status = Unconfirmed
					go orderBuffer(order, channels.OrderUpdate)
					break
				}
				if allOrders[order.Floor].Status != Confirmed {
					allOrders[order.Floor].Status = Confirmed
					go timingOrder(allOrders[order.Floor], channels.OrderUpdate, order.Cost[findIdWithLowestCost(order)]+DOOR_OPEN_TIME*NUMBER_OF_FLOORS)
				}

			case Mine:
				fmt.Println("*** STATUS Mine: \t", order.Floor)
				if allOrders[order.Floor].Status > Mine {
					fmt.Println("*** order with status Mine CANCELLED: \t", order.Floor)
					break
				}

				if order.TimedOut == true && order.CabOrder == false {
					fmt.Println("** order with status Mine TIMEDOUT: \t", order.Floor)
					order.Status = Confirmed
					allOrders[order.Floor].Status = Confirmed
					go orderBuffer(order, orderToPeers)
					break
				} else if order.TimedOut == true {
					break
				}

				if order.DirectionUp == true {
					allOrders[order.Floor].DirectionUp = true
				}
				if order.DirectionDown == true {
					allOrders[order.Floor].DirectionDown = true
				}
				go orderBuffer(allOrders[order.Floor], orderOut)
				fmt.Println("*** ORDER SENT TO FSM: \t", order.Floor)
				go timingOrder(order, channels.OrderUpdate, order.Cost[ELEVATOR_ID]*3+5)
				break

			case Done:
				fmt.Println("****** ORDER DONE: \t", order.Floor)
				if order.FromId == ELEVATOR_ID {
					go orderBuffer(order, orderToPeers)
					elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, false)
					order.CabOrder = false
				}
				elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, false)
				elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, false)

				order.Status = NotActive
				order.DirectionUp = false
				order.DirectionDown = false

				order.TimedOut = false
				order.Timestamp = time.Now().Unix()
				for elevatorNumber := 0; elevatorNumber < NUMBER_OF_ELEVATORS; elevatorNumber++ {
					order.Cost[elevatorNumber] = MAXCOST
				}
				allOrders[order.Floor] = order
				break
			}

		case elevatorState = <-getElevatorState:
			/* elevatorIsImmobile is a variable to only trigger this if once every time the elevator becomes immobile */
			if elevatorState.Immobile && !elevatorIsImmobile {
				for floor := 0; floor < NUMBER_OF_FLOORS; floor++ {
					if allOrders[floor].Status == Mine {
						allOrders[floor].Status = Confirmed
						allOrders[floor].TimedOut = true
						go orderBuffer(allOrders[floor], orderToPeers)
						allOrders[floor].DirectionUp = false
						allOrders[floor].DirectionDown = false
					}
				}
			}
			elevatorIsImmobile = elevatorState.Immobile

		default:
		}
	}
}
