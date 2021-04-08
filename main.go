package main

import (
	//"./elevio"
	//"./fsm"
	"./timer"
)

func main() {

	TimedOut := make(chan bool)
	go timer.DoorTimer(3, TimedOut)

	if <-TimedOut {
		println("finished timer")
	}

}
