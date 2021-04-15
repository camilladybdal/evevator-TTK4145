package orderDistributor

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	"../network/bcast"
	. "../types"
	. "../costfnc"
)

func orderNetworkCommunication(orderToNetwork <-chan Order, orderFromNetwork chan<- Order) {
	port := Port
	networkTransmit := make(chan Order)
	networkRecieve := make(chan Order)

	go bcast.Transmitter(port, networkTransmit)
	go bcast.Receiver(port, networkRecieve)

	for {
		select {
		case order := <-orderToNetwork:
			fmt.Println("Order sent to network")
			/*
			redundancy := 1
			for redundancy > 0 {
				networkTransmit <- order
				time.Sleep(10 * time.Millisecond)
				redundancy--
			}
			*/
			networkTransmit <- order

		case order := <-networkRecieve:
			fmt.Println("Order recv from network")
			go orderBuffer(order, orderFromNetwork)
		}
	}
}

func orderTimer(order Order, timedOut chan<- Order, duration int) {

	// Quick fix! NEED TO CHANGE
	for duration > 0 {
		//fmt.Println(duration - 1)
		time.Sleep(time.Second)
		duration--
	}
	order.TimedOut = true
	fmt.Println("Order timer expired")
	timedOut <- order
}

func orderBuffer(order Order, orderIn chan<- Order) {
	fmt.Println("Order in buffer")
	orderIn <- order
}

func pollOrders(orderIn chan Order) {
	newButtonEvent := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(newButtonEvent)

	for {
		select {
		case buttonEvent := <-newButtonEvent:
			fmt.Println("newButtonEvent")
			var newOrder Order
			newOrder.Floor = buttonEvent.Floor
			fmt.Println("Button pressed, F: ", newOrder.Floor)
			buttonType := buttonEvent.Button
			newOrder.DirectionUp = (buttonType == elevio.BT_HallUp)
			newOrder.DirectionDown = (buttonType == elevio.BT_HallDown)
			newOrder.CabOrder = (buttonType == elevio.BT_Cab)

			for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
				newOrder.Cost[elevatorNumber] = MaxCost
			}

			newOrder.Status = 1
			newOrder.TimedOut = false
			go orderBuffer(newOrder, orderIn)
		}
	}
}

func orderFindIdWithLowestCost(order Order) (int) {
	lowestCostId := ElevatorId
	for elevator := 0; elevator < NumberOfElevators; elevator++ {
		if order.Cost[elevator] < order.Cost[lowestCostId] {
			lowestCostId = elevator
		} 
	}
	fmt.Println("Lowest cost id: ", lowestCostId)
	return lowestCostId
}


func OrderDistributor(orderOut chan<- Order, orderIn chan Order, getElevatorState <-chan Elevator) {
	fmt.Println("Starting OrderDistributor...")
	var queue [NumberOfFloors]Order
	go pollOrders(orderIn)
	orderToNetworkChannel := make(chan Order)

	go orderNetworkCommunication(orderToNetworkChannel, orderIn)

	var elevatorState Elevator

	for floor := 0; floor < NumberOfFloors; floor++ {
		queue[floor].Floor = floor
		queue[floor].DirectionUp = false
		queue[floor].DirectionDown = true
		queue[floor].CabOrder = false
		for elevator := 0; elevator < NumberOfElevators; elevator++ {
			queue[floor].Cost[elevator] = MaxCost
		}
		queue[floor].Status = NoActiveOrder
		queue[floor].TimedOut = false
	}
	for {
		select {
		// Order pipeline
		case order := <-orderIn:
			fmt.Println("Reading order...")
			if queue[order.Floor].Status == NoActiveOrder && order.TimedOut {
				fmt.Println("Order early break senor, F: ", order.Floor)
				break
			}

			switch order.Status {
			case NoActiveOrder:
			case WaitingForCost:
				fmt.Println("Status is Waiting for cost, F: ", order.Floor)

				if order.CabOrder == true {
					order.Cost[ElevatorId] = Costfunction(elevatorState, order) // Bruk costfunction
					order.Status = Confirmed
					order.CabOrder = false
					orderToNetworkChannel <- order
					order.CabOrder = true
					order.Status = Mine
					queue[order.Floor] = order
					go orderBuffer(order, orderIn)
					break
				}
				if queue[order.Floor].Status > WaitingForCost {
					fmt.Println("Already higher status than Waiting for cost, F:", order.Floor)
					break
				}
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
					fmt.Println("Adding own cost")
					queue[order.Floor].Cost[ElevatorId] = Costfunction(elevatorState, order)
					orderToNetworkChannel <- queue[order.Floor]
					go orderTimer(order, orderIn, 1)
				}

				allCostsPresent := true
				for elevator := 0; elevator < NumberOfElevators; elevator++ {
					if queue[order.Floor].Cost[elevator] == MaxCost {
						allCostsPresent = false
					}
				}

				if allCostsPresent || order.TimedOut {
					fmt.Println("WFC ACP or order TO")
					queue[order.Floor].Status = Unconfirmed
					go orderBuffer(queue[order.Floor], orderIn)
					orderToNetworkChannel <- queue[order.Floor]
				}
				break

			case Unconfirmed:
				fmt.Println("Status is Unconfirmed, F: ", order.Floor)
				if queue[order.Floor].Status > Unconfirmed {
					fmt.Println("Already higher status than Unconfirmed, F: ", order.Floor)
					break
				}
				if order.TimedOut == true {
					queue[order.Floor].Cost[orderFindIdWithLowestCost(order)] = MaxCost
					go orderBuffer(queue[order.Floor], orderIn)
				}

				if orderFindIdWithLowestCost(order) == ElevatorId {
					fmt.Println("Has lowest cost")
					queue[order.Floor].Status = Confirmed
					orderToNetworkChannel <- queue[order.Floor]
					queue[order.Floor].Status = Mine
					go orderBuffer(queue[order.Floor], orderIn)
				} else {
					go orderTimer(queue[order.Floor], orderIn, 1)
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

				fmt.Println("Status is Confirmed, F: ", order.Floor)
				if queue[order.Floor].Status > Confirmed {
					fmt.Println("Already higher status than Confirmed, F: ", order.Floor)
					break
				}

				if order.TimedOut == true {
					queue[order.Floor].Status = Unconfirmed
					order.Status = Unconfirmed
					go orderBuffer(order, orderIn)
					break
				}

				queue[order.Floor] = order
				go orderTimer(order, orderIn, order.Cost[orderFindIdWithLowestCost(order)]*2) // Må endres til et uttrykk med costen
				break // Hva skjer hvis alle har MaxCost?

			case Mine:
				fmt.Println("Status is Mine, F: ", order.Floor)
				if queue[order.Floor].Status > Mine {
					fmt.Println("Order with status Mine cancelled, F: ", order.Floor)
					break
				}

				if order.TimedOut == true && order.CabOrder == false {
					fmt.Println("Order with status Mine has Timed out, F:", order.Floor)
					order.Status = Confirmed
					queue[order.Floor].Status = Confirmed
					orderToNetworkChannel <- order
					break
				} else if order.TimedOut == true {
					fmt.Println("Error: Could not expedite Caborder!!")
					break
				}

				// send til fsm
				orderOut <- queue[order.Floor]
				fmt.Println("Order sent to FSM, F: ", order.Floor)
				go orderTimer(order, orderIn, order.Cost[ElevatorId]*2) // Må også endres
				break

			case Done:
				fmt.Println("Status is Done, F: ", order.Floor)
				order.Status = NoActiveOrder
				order.DirectionUp = false
				order.DirectionDown = false
				order.CabOrder = false
				order.TimedOut = false
				for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
					order.Cost[elevatorNumber] = MaxCost
				}
				queue[order.Floor] = order
				orderToNetworkChannel <- order
				break
			}

		case elevatorState = <- getElevatorState:
			break

		default:
			
		}
	}
}
