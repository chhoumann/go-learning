package main

import (
	"fmt"
	"time"
)

// === DONE CHANNEL ===
func doWork(done <-chan bool) {
	counter := 0

	for {
		select {
		case <-done:
			// done working
			fmt.Printf("Counted to %d\n", counter)
			return
		default:
			// working
			counter += 1
		}
	}
}

func doneChannelPattern() {
	done := make(chan bool)

	go doWork(done)

	time.Sleep(3 * time.Second)
	done <- true

	close(done)
}
