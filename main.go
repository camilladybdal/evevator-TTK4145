package main

import (
	//"./elevio"
	"./fsm"
	"./timer"
)

func main() {
	fsm.InitFSM(4)

	TimedOut := make(chan bool)
	go timer.DoorTimer(3, TimedOut)

	if <-TimedOut {
		println("finished timer")
	}

}
