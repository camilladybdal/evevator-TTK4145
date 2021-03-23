package timer

import (
	"fmt"
	"time"
)

func Timer(interval time.Duration, id int, stopTimer <-chan bool, timedOut chan<- int) {

	timer := time.NewTimer(interval * time.Second) //will send the time on the channel after each tick

	for {
		select {
		case <-stopTimer:
			fmt.Println("Finished in time")

			//draining channel
			if !timer.Stop() {
				<-timer.C
			}
			return

		case <-timer.C:
			fmt.Println("Timed out")
			timedOut <- id
		}
	}
}
