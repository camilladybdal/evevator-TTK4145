package timer

import (
	"time"
)

func StoppableTimer(interval time.Duration, id int, stopTimer <-chan bool, timedOut chan<- int) {
	timer := time.NewTimer(interval * time.Second)

	for {
		select {
		case <-stopTimer:
			if !timer.Stop() {
				<-timer.C
			}
			return

		case <-timer.C:
			timedOut <- id
		}
	}
}

func ResetableTimer(duration time.Duration, newCountdownTime <-chan time.Duration, timedOut chan<- bool) {
	timer := time.NewTimer(duration * time.Second)

	for {
		select {
		case newTime := <-newCountdownTime:
			timer.Reset(newTime * time.Second)
		case <-timer.C:
			timedOut <- true
		}
	}
}
