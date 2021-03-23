package orderDistributor

// imports

// Constants
const (
	NumberOfElevators = 3 // Need better implemantation (config fil?)
	NumberOfFloors    = 4 // also config?
)

// Structures
type Order struct {
	Floor     int
	Direction [2]bool
	Cost      [NumberOfElevators]int
	Status    int // 0: No active order , 1: waiting for cost, 2: unconfirmed, 3: confirmed, 4: mine, 5: done
	Deadline  int // Time? or Id?
}

// Button struct?

// Functions

// FSM_ORDER_IN ONLY ONE SPACE IN CHANNEL
// orderIn kan få ordre fra både nettverket og elevio?
func OrderDistributor(orderOut chan<- Order, orderExpedited <-chan Order, orderIn chan Order) {
	var queue [NumberOfFloors]Order

	for floor := 0; floor < NumberOfFloors; floor++ {
		queue[floor].Floor = floor
	}

	for {
		select {
		case order := <-orderIn:
			switch order.Status {
			case 0:
				//Kanskje noe?

			case 1:
				if queue[order.Floor].Status > 1 {
					break
				}
				// Set queue status = 1
				// If own cost not attached, Calculate, add and share (start timer?)
				// else: update queue with new costs

				// If all costs present queue order status += 1 (or if timer timed out)
				orderIn <- order
			case 2:
				// If this has lowest cost:
				// status = 3, share on network, status = 4
				orderIn <- order

				// else:
				// Set timer?

			case 3:
				// if not 4 or 5
				// Set status to 3
				// Set finish timer

			case 4:
				orderOut <- order
				// Set timer

			case 5:
				// Clear order in queue
				// Share on network
				// Set status to 0

			}

		}
	}

}
