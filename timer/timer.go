package timer

import (
	"fmt"
	"time"
)

type TimerStruct struct {
	Interval time.Duration
	Id       int
	Done     bool
}

var ticker *time.Ticker

/*
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		time.Sleep(10 * time.Second)
		done <- true
	}()
	for {
		select {
		case <-done:
			fmt.Println("Done!")
			return
		case t := <-ticker.C:
			fmt.Println("Current time: ", t)
		}
	}
*/

/* How to use:
timeout := make(chan bool)
done := make(chan bool)

starting timer:
go Timer(interval, done, orderTimeout)

when elevator is finished = done <- true
while timout = false, whatever finished in time


 done <-chan bool
*/

func Timer(timerinput <-chan TimerStruct, timeout chan<- int) {

	info := <-timerinput

	ticker = time.NewTicker(info.Interval * time.Second) //will send the time on the channel after each tick

	for {
		select {
		case t := <-timerinput:
			if t.Done {
				fmt.Println("Finished in time")
			}

		case t := <-ticker.C:
			fmt.Println(t, "Timed out")
			timeout <- info.Id
		}
	}
}
