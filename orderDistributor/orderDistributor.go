package orderDistributor

// imports
import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	"../network/bcast"
	. "../types"
	//"../costfnc"
)

// Slå sammen netverksfunksjonene

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
			redundancy := 1
			for redundancy > 0 {
				networkTransmit <- order
				time.Sleep(10 * time.Millisecond)
				redundancy--
			}
		case order := <-networkRecieve:
			fmt.Println("Order recv from network")
			orderFromNetwork <- order
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
	lowestCostId = ElevatorId
	for elevator := 0; elevator < NumberOfElevators; elevator++ {
		if order.Cost[elevator] < order.Cost[lowestCostId] {
			lowestCostId = elevator
		} 
	}
	return elevator
}


func OrderDistributor(orderOut chan<- Order, orderIn chan Order, getElevatorState <-chan Elevator) {
	fmt.Println("Starting OrderDistributor...")
	var queue [NumberOfFloors]Order
	go pollOrders(orderIn)
	orderToNetworkChannel := make(chan Order)

	go orderNetworkCommunication(orderToNetworkChannel, orderIn)

	//var elevatorState Elevator

	for floor := 0; floor < NumberOfFloors; floor++ {
		queue[floor].Floor = floor
	}
	for {
		select {
		// Order pipeline
		case order := <-orderIn:
			fmt.Println("Reading order...")
			switch order.Status {
			case NoActiveOrder:
			case WaitingForCost:
				fmt.Println("Status is Waiting for cost, F: ", order.Floor)

				if order.CabOrder == true {
					order.Cost[ElevatorId] = 5 // Bruk costfunction
					order.Status = Confirmed
					order.CabOrder = false
					orderToNetwork <- order
					order.CabOrder = true
					order.Status = Mine
					queue[order.Floor] = order
				}
				if queue[order.Floor].Status > WaitingForCost {
					fmt.Println("Already higher status than Waiting for cost, F:" order.Floor)
					break
				}
				queue[order.Floor].DirectionUp |= order.DirectionUp
				queue[order.Floor].DirectionDown |= order.DirectionDown

				for elevator := 0; elevator < NumberOfElevators; elevator++ {
					if order.Cost[elevator] != MaxCost {
						queue[order.Floor].Cost = order.Cost[elevator] // Sjekke om det ikke oppstår uenigheter
					}
				}
				if queue[order.Floor].Cost[ElevatorId] == MaxCost {
					queue[order.Floor].Cost[ElevatorId] = 5 // Costfnc
					orderToNetwork <- queue[order.Floor]
					orderTimer(order, orderIn, 1)
				}

				allCostsPresent := true
				for elevator := 0; elevator < NumberOfElevators; elevator++ {
					if queue[order.Floor].Cost[elevator] == MaxCost {
						allCostsPresent = false
					}
				}

				if allCostsPresent || order.TimedOut {
					queue[order.Floor].Status = Unconfirmed
					orderBuffer(queue[order.Floor], orderIn)
					orderToNetwork <- queue[order.Floor]
				}
				break
				/*
				if order.Cost[ElevatorId] == MaxCost {
					cost := 4 // add costfnc
					order.Cost[ElevatorId] = cost
					order.TimedOut = false
					orderToNetworkChannel <- order
					queue[order.Floor] = order

					go orderTimer(order, orderIn, 2)
				}

				// Not sure if this is the best solution
				allCostsPresent := true
				for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
					if order.Cost[elevatorNumber] == MaxCost {
						allCostsPresent = false
					}
				}
				// Hvis timeout, må sende ordre fra queue
				if allCostsPresent || order.TimedOut {
					order.Status = Unconfirmed
					queue[order.Floor] = order
					order.TimedOut = false
					go orderBuffer(order, orderIn)
				}
				break
				*/

			case Unconfirmed:
				fmt.Println("Status is Unconfirmed, F: ", order.Floor)
				if queue[order.Floor].Status > Unconfirmed {
					fmt.Println("Already higher status than Unconfirmed, F: ", order.Floor)
					break
				}
				if order.TimedOut == true {
					queue[order.Floor].Cost[orderFindIdWithLowestCost(order)] = MaxCost
					orderBuffer(queue[order.Floor], orderIn)
				}

				if orderFindIdWithLowestCost(order) == ElevatorId {
					queue[order.Floor].Status = Confirmed
					orderToNetwork <- queue[order.Floor]
					queue[order.Floor].Status = Mine
					orderBuffer(queue[order.Floor], orderIn)
				} else {
					orderTimer(queue[order.Floor], orderIn, 1)
				}
				break

			case Confirmed:
				// Sette på lys

				if order.DirectionUp == true {
					elevto.SetButtonLamp(elevio.BT_HallUp, order.Floor, true)
				}
				if order.DirectionDown == true {
					elevto.SetButtonLamp(elevio.BT_HallDown, order.Floor, true)
				}


				fmt.Println("Status is Confirmed, F: ", order.Floor)
				if queue[order.Floor].Status > Confirmed {
					fmt.Println("Already higher status than Confirmed, F: " order.Floor)
					break
				}
				if order.TimedOut == true {
					order.Status = Mine
					order.TimedOut = false
					go orderBuffer(order, orderIn)
					break
				}

				order.TimedOut = false
				queue[order.Floor] = order
				go orderTimer(order, orderIn, (10 + ElevatorId)) // Må endres til et uttrykk med costen
				break

			case Mine:
				fmt.Println("Status is Mine")
				if queue[order.Floor].Status > Mine || (queue[order.Floor].Status < Mine && order.TimedOut == true) {
					fmt.Println("Order with status Mine cancelled")
					break
				}
				if order.TimedOut == true {
					fmt.Println("Order with status Mine has Timed out")
					order.Cost[ElevatorId] = MaxCost
					order.Status = Unconfirmed
					orderToNetworkChannel <- order
					break
				}

				go orderTimer(order, orderIn, 5) // Må også endres
				break

			case Done:
				fmt.Println("Status is Done")
				order.Status = NoActiveOrder
				order.DirectionUp = false
				order.DirectionDown = false
				order.TimedOut = false
				for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
					order.Cost[elevatorNumber] = MaxCost
				}
				queue[order.Floor] = order
				orderToNetworkChannel <- order
				break
			}
			break
		default:

			// Getting the latest elevatorState
			/*
				case elevatorState = <- getElevatorState:
					break
			*/
		}
	}
}
