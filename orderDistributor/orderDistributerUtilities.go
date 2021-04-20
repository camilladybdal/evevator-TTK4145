package orderDistributor

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	. "../types"
)

func sendOrderMultipleTimes(order Order, orderToTransmitter chan<- Order) {
	redundancy := 5
	for redundancy > 0 {
		orderToTransmitter <- order
		time.Sleep(20 * time.Millisecond)
		redundancy--
	}
}

func orderCommunication(orderToTransmitter chan<- Order, orderFromReciever <-chan Order, orderToPeers <-chan Order, orderFromPeers chan<- Order) {

	for {
		select {
		case order := <-orderToPeers:
			order.FromId = ELEVATOR_ID
			order.CabOrder = false

			go sendOrderMultipleTimes(order, orderToTransmitter)
			if order.FromId == ELEVATOR_ID {
				break
			}

		case order := <-orderFromReciever:
			if order.FromId == ELEVATOR_ID {
				break
			}
			go orderBuffer(order, orderFromPeers)
		}
	}
}

func timingOrder(order Order, timedOut chan<- Order, duration int) {
	order.Timestamp = time.Now().Unix()
	for duration > 0 {
		time.Sleep(time.Second)
		duration--
	}
	order.TimedOut = true
	order.FromId = ELEVATOR_ID
	fmt.Println("*** order timer expired: \t", order.Floor, order.Status)
	timedOut <- order
}

func orderBuffer(order Order, bufferTo chan<- Order) {
	bufferTo <- order
	return
}

func pollOrders(orderIn chan Order, newButtonEvent <-chan elevio.ButtonEvent) {
	for {
		select {
		case buttonEvent := <-newButtonEvent:
			var newOrder Order
			newOrder.Floor = buttonEvent.Floor
			fmt.Println("*** BUTTON pressed: \t", newOrder.Floor)
			buttonType := buttonEvent.Button
			newOrder.DirectionUp = (buttonType == elevio.BT_HallUp)
			newOrder.DirectionDown = (buttonType == elevio.BT_HallDown)
			newOrder.CabOrder = (buttonType == elevio.BT_Cab)

			for elevatorNumber := 0; elevatorNumber < NUMBER_OF_ELEVATORS; elevatorNumber++ {
				newOrder.Cost[elevatorNumber] = MAXCOST
			}
			newOrder.Status = WaitingForCost
			newOrder.TimedOut = false
			go orderBuffer(newOrder, orderIn)
		}
	}
}

func findIdWithLowestCost(order Order) int {
	lowestCostId := ELEVATOR_ID
	for elevator := 0; elevator < NUMBER_OF_ELEVATORS; elevator++ {
		if 10*order.Cost[elevator]+elevator < 10*order.Cost[lowestCostId]+lowestCostId {
			lowestCostId = elevator
		}
	}
	if order.Cost[lowestCostId] == MAXCOST {
		lowestCostId = ELEVATOR_ID
		fmt.Println("****** all elevators MAXCOST: \t", order.Floor)

	}
	return lowestCostId
}

// ???
func drainChannels(orderIn chan Order, startDraining <-chan bool) {
	drain := false
	for {
		select {
		case drain = <-startDraining:

		}
		if drain == true {
			fmt.Println("*** DRAINING!!!------------------------------------------------------------------")
			select {
			case <-orderIn:
			}
		}
	}
}
