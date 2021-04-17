package timer

import (
	//"fmt"
	"time"
)

func StoppableTimer(interval time.Duration, id int, stopTimer <-chan bool, timedOut chan<- int) {

	timer := time.NewTimer(interval * time.Second) //will send the time on the channel after each tick

	for {
		select {
		case <-stopTimer:
			//fmt.Println("Finished in time")

			//draining channel
			if !timer.Stop() {
				<-timer.C
			}
			return

		case <-timer.C:
			//fmt.Println("Timed out")
			timedOut <- id
		}
	}
}


func ResetableTimer(duration time.Duration, newCountdownTime <-chan time.Duration, timedOut chan<- bool) {
	timer := time.NewTimer(duration * time.Second)

	for {
		select {
		case newTime := <- newCountdownTime:
			timer.Reset(newTime * time.Second)
		case <-timer.C:
			timedOut <- true
		}
	}
}