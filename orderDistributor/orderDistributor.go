package orderDistributor

// imports
import (
	"fmt"
	"time"

	"../elevio"
	. "../types"
	"../network/bcast"
	"../config"
	//"../costfnc"
)

func orderToNetwork(orderToNetwork <-chan Order) {
	port := config.Port
	networkTransmit := make(chan Order)

	go bcast.Transmitter(port, networkTransmit)

	for {
		select {
		case order := <- orderToNetwork:
			fmt.Println("Order sent to network")
			redundancy := 5
			order.Status = Unconfirmed
			for redundancy > 0 {
				networkTransmit <- order
				time.Sleep(10*time.Millisecond)
				redundancy--
			}
			
		}
	}
}

func orderFromNetwork(orderFromNetwork chan<- Order) {
	port := config.Port
	networkRecieve := make(chan Order)

	go bcast.Receiver(port, networkRecieve)

	for {
		select {
		case orderFromNetwork <- <- networkRecieve:

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
	timedOut <- order
}

func orderBuffer(order Order, orderIn chan<- Order) {
	fmt.Println("Order in buffer")
	orderIn <- order
	fmt.Println("Order sent to PL")
}

func pollOrders(orderIn chan Order) {
	fmt.Println("Polling orders...")
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

func OrderDistributor(orderOut chan<- Order, orderIn chan Order, getElevatorState <-chan Elevator) {
	fmt.Println("Starting OD...")
	var queue [NumberOfFloors]Order
	go pollOrders(orderIn)
	orderToNetworkChannel := make(chan Order)
	go orderToNetwork(orderToNetworkChannel)
	go orderFromNetwork(orderIn)

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
				fmt.Println("Status is Waiting for cost")
				if queue[order.Floor].Status > WaitingForCost {
					fmt.Println("Already higher status than Waiting for cost")
					break
				}
				if order.Cost[ElevatorId] == MaxCost {
					// TODO: Ask for elevator state and calculate cost using cost function
					cost := 5
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
				if allCostsPresent || order.TimedOut {
					order.Status = Unconfirmed
					queue[order.Floor] = order
					order.TimedOut = false
					go orderBuffer(order, orderIn)
				}
				break

			case Unconfirmed:
				fmt.Println("Status is Unconfirmed")
				if queue[order.Floor].Status > Unconfirmed {
					fmt.Println("Already higher status than Unconfirmed")
					break
				}
				if order.TimedOut == true {
					order.Status = Mine
					queue[order.Floor] = order
					order.TimedOut = false
					go orderBuffer(order, orderIn)
				}

				hasLowestCost := true
				for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
					if order.Cost[elevatorNumber]*10+elevatorNumber < order.Cost[ElevatorId]*10+ElevatorId {
						hasLowestCost = false
					}
				}
				if hasLowestCost {
					order.Status = Confirmed
					// TODO share on network
					order.Status = Mine
					queue[order.Floor] = order
					go orderBuffer(order, orderIn)
				} else {
					go orderTimer(order, orderIn, 1)
				}
				break

			case Confirmed:
				fmt.Println("Status is Confirmed")
				if queue[order.Floor].Status > Confirmed {
					fmt.Println("Already higher status than Confirmed")
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
					// TODO share on network
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
				// TODO Share on network
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
