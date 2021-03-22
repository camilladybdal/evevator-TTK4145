package orderDistributor

// imports

// Constants
const (
	NumberOfElevators = 3 // Need better implemantation (config fil?)
)

// Structures
type Order struct {
	Floor    int
	Cost     [NumberOfElevators]int
	Status   int // 0: , 1: waiting for cost, 2: unconfirmed, 3: confirmed, 4: mine, 5: done
	Deadline int // Time? or Id?
}

// Button struct?

// Functions

// FSM_ORDER_IN ONLY ONE SPACE IN CHANNEL
// orderIn kan få ordre fra både nettverket og elevio?
func OrderDistributor(orderOut chan<- Order, orderExpedited <-chan Order, orderIn <-chan Order) {

	// Polle etter knapper her?

	for {
		select {
		case order := <-orderIn:
			switch order.Status {
			case 0:
				// Calculate and add cost
				// Update local data
				// Set reponse timer
				// Set Status to 1 if *something*
				// Share on network

			case 1:
				// Calculate and add cost if not already present
				//

			}

		}
	}

}
