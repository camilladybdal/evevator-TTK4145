package orderDistributor

// imports
import (
	"time"
	"../FSM"
	"../types"

)

func orderTimer(order Order, timedOut chan<- Order, duration int) {

	// Quick fix! NEED TO CHANGE
	for duration > 0 {
		time.Sleep(time.Second)
		duration--
	}
	order.TimedOut = true
	timedOut <- order
}

// orderIn kan få ordre fra både nettverket og elevio
func OrderDistributor(orderOut chan<- Order, orderExpedited <-chan Order, orderIn chan Order, getElevatorState <-chan types.Elevator) {
	var queue [NumberOfFloors]Order
	var elevatorState types.Elevator

	for floor := 0; floor < NumberOfFloors; floor++ {
		queue[floor].Floor = floor
	}

	for {
		select {
			// Order pipeline
		case order := <-orderIn:
			switch order.Status {
			case noActiveOrder:
				// Kanskje noe?
				// Log some sort of error?
				break

			case waitingForCost:
				if queue[order.Floor].Status > waitingForCost {
					break
				}
				// If own cost not attached, Calculate, add and share (start timer?)
				// else: update queue with new costs
				if order.Cost[elevatorId] == maxCost {
					// TODO: Ask for elevator state and calculate cost using cost function
					cost = costfnc.Costfunction(elevatorState, order)
					order.Cost[elevatorId] = cost
					order.TimedOut = false
					// TODO: Share order on network
					queue[order.Floor] = order
					go orderTimer(order, orderIn, 3)
				}

				// Not sure if this is the best solution
				allCostsPresent := true
				for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
					if order.Cost[elevatorNumber] == maxCost {
						allCostsPresent = false
					}
				}
				if allCostsPresent || order.TimedOut {
					order.Status = unconfirmed
					queue[order.Floor] = order
					orderIn <- order
				}
				break

			case unconfirmed:
				if queue[order.Floor].Status > unconfirmed {
					break
				}

				hasLowestCost := true
				for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
					if order.Cost[elevatorNumber]*10+elevatorNumber < order.Cost[elevatorId]*10+elevatorId {
						hasLowestCost = false
					}
				}
				if hasLowestCost {
					order.Status = confirmed
					// TODO share on network
					order.Status = mine
					queue[order.Floor] = order
					orderIn <- order
				} else {
					go orderTimer(order, orderIn, 3)
				}
				break

			case confirmed:
				if queue[order.Floor].Status > confirmed {
					break
				}
				if order.TimedOut == true {
					order.Status = mine
					orderIn <- order
					break
				}

				order.TimedOut = false
				queue[order.Floor] = order
				go orderTimer(order, orderIn, 10) // Må endres til et uttrykk med costen
				break

			case mine:
				if queue[order.Floor].Status > mine {
					break
				}
				if order.TimedOut == true {
					order.Cost[elevatorId] = maxCost
					order.Status = unconfirmed
					// TODO share on network
					break
				}
				orderOut <- order

				go orderTimer(order, orderIn, 10) // Må også endres
				break

			case done:
				// Clear order in queue
				order.Status = noActiveOrder
				order.DirectionUp = false
				order.DirectionDown = false
				order.TimedOut = false
				for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
					order.Cost[elevatorNumber] = maxCost
				}
				queue[order.Floor] = order
				// TODO Share on network
				break
			}
			break

		// Getting the latest elevatorState
		case elevatorState = <- getElevatorState:
			break

		}
	}
}