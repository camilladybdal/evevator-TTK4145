package orderDistributor

// imports
import (
	"time"
	"fmt"
	."../types"
	"../elevio"
	//"../costfnc"

)

func orderTimer(order Order, timedOut chan<- Order, duration int) {

	// Quick fix! NEED TO CHANGE
	for duration > 0 {
		fmt.Println(duration-1)
		time.Sleep(time.Second)
		duration--
	}
	order.TimedOut = true
	timedOut <- order
}

func orderBuffer(order Order, orderIn chan<- Order) {
	orderIn <- order
}

func pollOrders(orderIn chan Order) {
	newButtonEvent := make(chan elevio.ButtonEvent)
	elevio.PollButtons(newButtonEvent)

	for {
		select {
		case buttonEvent := <- newButtonEvent:
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
	var queue [NumberOfFloors]Order
	go PollOrders(orderIn)
	//var elevatorState Elevator

	for floor := 0; floor < NumberOfFloors; floor++ {
		queue[floor].Floor = floor
	}
	for {
		select {
			// Order pipeline
		case order := <-orderIn:
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
					// TODO: Share order on network
					queue[order.Floor] = order
					
					go orderTimer(order, orderIn, 2)
					fmt.Println("Starting timer in WFC")
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
					fmt.Println("!!")
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
				go orderTimer(order, orderIn, (10+ElevatorId)) // Må endres til et uttrykk med costen
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
				orderOut <- order

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