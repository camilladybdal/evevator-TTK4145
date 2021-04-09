package orderDistributor

// imports
import (
	"time"
)

// Constants
const (
	NumberOfElevators = 3 // Need better implemantation (config fil?)
	NumberOfFloors    = 4 // also config?
	maxCost = 999999999
)

// Structures
type Order struct {
	Floor     int
	DirectionUp bool
	DirectionDown bool
	Cost      [NumberOfElevators]int
	Status    int // 0: No active order , 1: waiting for cost, 2: unconfirmed, 3: confirmed, 4: mine, 5: done
	TimedOut  int // Time? or Id?
}

// Button struct?

// Functions
func orderTimer(order Order, timedOut chan<- Order, duration int) {

	time.Sleep(duration*time.Second)
	order.TimedOut = true
	timedOut <- order
}

// orderIn kan få ordre fra både nettverket og elevio?
func OrderDistributor(orderOut chan<- Order, orderExpedited <-chan Order, orderIn chan Order) {
	var queue [NumberOfFloors]Order

	for floor := 0; floor < NumberOfFloors; floor++ {
		queue[floor].Floor = floor
	}

	for {
		select {
			// Order pipeline
		case order := <-orderIn:
			switch order.Status {
			case 0:
				//Kanskje noe?
				// Log some sort of error?
				break

			case 1:
				if queue[order.Floor].Status > 1 {
					break
				}
				// If own cost not attached, Calculate, add and share (start timer?)
				// else: update queue with new costs
				if order.Cost[elevatorId] == maxCost {
					// TODO: Ask for elevator state and calculate cost using cost function
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
					order.Status += 1
					queue[order.Floor] = order
					orderIn <- order
				}
				break

			case 2:
				if queue[order.Floor].Status > 2 {
					break
				}

				hasLowestCost := true
				for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
					if order.Cost[elevatorNumber]*10+elevatorNumber < order.Cost[elevatorId]*10+elevatorId {
						hasLowestCost = false
					}
				}
				if hasLowestCost {
					order.Status = 3
					// TODO share on network
					order.Status = 4
					queue[order.Floor] = order
					orderIn <- order
				} else {
					go orderTimer(order, orderIn, 3)
				}
				break

			case 3:
				if queue[order.Floor].Status > 3 {
					break
				}
				if order.TimedOut = true {
					order.Status = 4
					orderIn <- order
					break
				}

				order.TimedOut = false
				queue[order.Floor] = order
				go orderTimer(order, orderIn, 10) // Må endres til et uttrykk med costen
				break
			case 4:
				if queue[order.Floor].Status > 4{
					break
				}
				if order.TimedOut = true {
					order.Cost[elevatorId] = maxCost
					order.Status = 3
					// TODO share on network
					break
				}
				orderOut <- order

				go orderTimer(order, orderIn, 10) // Må også endres
				break

			case 5:
				// Clear order in queue
				order.Status = 0
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
		}
	}
}