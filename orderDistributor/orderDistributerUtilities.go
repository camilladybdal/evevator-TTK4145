package orderDistributor

import (
	"time"
	"fmt"
	. "../config"
	"../elevio"
	"../network/bcast"
	. "../types"
	//. "../costfnc"
)
func orderNetworkResending(order Order, orderToNetwork chan<- Order) {
	redundancy := 5
	for redundancy > 0 {
		orderToNetwork <- order
		time.Sleep(20 * time.Millisecond)
		redundancy--
	}
}

func orderNetworkCommunication(orderToNetwork <-chan Order, orderFromNetwork chan<- Order) {
	port := Port
	networkTransmit := make(chan Order)
	networkRecieve := make(chan Order)

	go bcast.Transmitter(port, networkTransmit)
	go bcast.Receiver(port, networkRecieve)

	for {
		select {
		case order := <-orderToNetwork:
			//fmt.Println("*** order sent to network: \t", order.Floor)
			order.FromId = ElevatorId
			order.CabOrder = false
			
			go orderNetworkResending(order, networkTransmit)

		case order := <-networkRecieve:
			if order.FromId == ElevatorId {
				//fmt.Println("*** read own order from network")
				break
			}
			//fmt.Println("*** order recv from network: \t", order.Floor)
			go orderBuffer(order, orderFromNetwork)
		}
	}
}

func orderTimer(order Order, timedOut chan<- Order, duration int) {
	order.Timestamp = time.Now().Unix()
	// Quick fix! NEED TO CHANGE
	for duration > 0 {
		//fmt.Println(duration - 1)
		time.Sleep(time.Second)
		duration--
	}
	order.TimedOut = true
	fmt.Println("*** order timer expired: \t", order.Floor, order.Status)
	timedOut <- order
}

func orderBuffer(order Order, bufferTo chan<- Order) {
	//fmt.Println("Order in buffer, F: ", order.Floor)
	bufferTo <- order
	return
}

func pollOrders(orderIn chan Order) {
	newButtonEvent := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(newButtonEvent)

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

			for elevatorNumber := 0; elevatorNumber < NumberOfElevators; elevatorNumber++ {
				newOrder.Cost[elevatorNumber] = MaxCost
			}
			newOrder.Status = WaitingForCost
			newOrder.TimedOut = false
			go orderBuffer(newOrder, orderIn)
		}
	}
}

func orderFindIdWithLowestCost(order Order) (int) {
	lowestCostId := ElevatorId
	for elevator := 0; elevator < NumberOfElevators; elevator++ {
		if 10*order.Cost[elevator]+elevator < 10*order.Cost[lowestCostId]+lowestCostId {
			lowestCostId = elevator //Tenk mer må dette etterpå
		}
	}
	if order.Cost[lowestCostId] == MaxCost {
		lowestCostId = ElevatorId
		fmt.Println("****** all elevators MAXCOST: \t", order.Floor)

	}
	//fmt.Println("*** lowest cost id: ", lowestCostId, " floor: \t", order.Floor)
	return lowestCostId
}

func drainChannels(orderIn chan Order, startDraining <-chan bool) {

	drain := false	

	for {
		select {
		case drain = <- startDraining:

		}
		if drain == true {
			//only do if possible (how?)
			fmt.Println("*** DRAINING!!!------------------------------------------------------------------")
			select {
			case <- orderIn:
			}
		}
	}
}