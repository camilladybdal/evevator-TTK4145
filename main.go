package main

import (
	//"./elevio"
	//"./fsm"
	//"./timer"
	"./orderDistributor"
	."./types"
	"fmt"
	"time"
)


func odTest() {

	orderFrom := make(chan Order)
	orderTo := make(chan Order,1)
	elevStateTo := make(chan Elevator)

	var elevatorState Elevator
	elevatorState.UpQueue = [4]int{0,0,0,0}
	elevatorState.DownQueue = [4]int{0,0,0,0}
	elevatorState.CurrentFloor = 0
	elevatorState.Direction = 0

	go orderDistributor.OrderDistributor(orderFrom, orderTo, elevStateTo)

	time.Sleep(time.Second)
	var order1 Order
	order1.Floor = 2
	order1.DirectionUp = false
	order1.DirectionDown = true
	order1.Status = WaitingForCost
	order1.TimedOut = false
	order1.Cost[0] = MaxCost
	order1.Cost[1] = MaxCost
	order1.Cost[2] = MaxCost
	orderTo <- order1
	order3 := order1
	order3.Cost[1] = 4
	orderTo <- order3
	//orderTo <- order1
	order2 := order1
	order2.Floor = 3
	orderTo <- order2


	for {
		select {

		case recv := <- orderFrom:
			fmt.Println("revcccc!")
			fmt.Println(recv.Floor)

			time.Sleep(2*time.Second)
			recv.Status = Done
			orderTo <- recv

		}
	}
}

func main() {

	go odTest()

	for{
	}

}