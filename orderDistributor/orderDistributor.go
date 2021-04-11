package orderDistributor

// imports
import (
	"time"
	"../types"
	"../costfnc"

)

func orderTimer(order types.Order, timedOut chan<- types.Order, duration int) {

	// Quick fix! NEED TO CHANGE
	for duration > 0 {
		time.Sleep(time.Second)
		duration--
	}
	order.TimedOut = true
	timedOut <- order
}

// orderIn kan få ordre fra både nettverket og elevio
func OrderDistributor(orderOut chan<- types.Order, orderIn <-chan types.Order, getElevatorState <-chan types.Elevator) {
	var queue [types.NumberOfFloors]types.Order
	var elevatorState types.Elevator

	for floor := 0; floor < types.NumberOfFloors; floor++ {
		queue[floor].Floor = floor
	}

	for {
		select {
			// Order pipeline
		case order := <-orderIn:
			switch order.Status {
			case types.NoActiveOrder:
				// Kanskje noe?
				// Log some sort of error?
				break

			case types.WaitingForCost:
				if queue[order.Floor].Status > types.WaitingForCost {
					break
				}
				// If own cost not attached, Calculate, add and share (start timer?)
				// else: update queue with new costs
				if order.Cost[types.ElevatorId] == types.MaxCost {
					// TODO: Ask for elevator state and calculate cost using cost function
					cost := costfnc.Costfunction(elevatorState, order)
					order.Cost[types.ElevatorId] = cost
					order.TimedOut = false
					// TODO: Share order on network
					queue[order.Floor] = order
					go orderTimer(order, orderIn, 3)
				}

				// Not sure if this is the best solution
				allCostsPresent := true
				for elevatorNumber := 0; elevatorNumber < types.NumberOfElevators; elevatorNumber++ {
					if order.Cost[elevatorNumber] == types.MaxCost {
						allCostsPresent = false
					}
				}
				if allCostsPresent || order.TimedOut {
					order.Status = types.Unconfirmed
					queue[order.Floor] = order
					orderIn <- order
				}
				break

			case types.Unconfirmed:
				if queue[order.Floor].Status > types.Unconfirmed {
					break
				}

				hasLowestCost := true
				for elevatorNumber := 0; elevatorNumber < types.NumberOfElevators; elevatorNumber++ {
					if order.Cost[elevatorNumber]*10+elevatorNumber < order.Cost[types.ElevatorId]*10+types.ElevatorId {
						hasLowestCost = false
					}
				}
				if hasLowestCost {
					order.Status = types.Uonfirmed
					// TODO share on network
					order.Status = types.Mine
					queue[order.Floor] = order
					orderIn <- order
				} else {
					go orderTimer(order, orderIn, 3)
				}
				break

			case types.Unconfirmed:
				if queue[order.Floor].Status > types.Confirmed {
					break
				}
				if order.TimedOut == true {
					order.Status = types.Mine
					orderIn <- order
					break
				}

				order.TimedOut = false
				queue[order.Floor] = order
				go orderTimer(order, orderIn, 10) // Må endres til et uttrykk med costen
				break

			case types.Mine:
				if queue[order.Floor].Status > types.Mine {
					break
				}
				if order.TimedOut == true {
					order.Cost[types.ElevatorId] = types.MaxCost
					order.Status = types.Unconfirmed
					// TODO share on network
					break
				}
				orderOut <- order

				go orderTimer(order, orderIn, 10) // Må også endres
				break

			case types.Done:
				// Clear order in queue
				order.Status = types.NoActiveOrder
				order.DirectionUp = false
				order.DirectionDown = false
				order.TimedOut = false
				for elevatorNumber := 0; elevatorNumber < types.NumberOfElevators; elevatorNumber++ {
					order.Cost[elevatorNumber] = types.MaxCost
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