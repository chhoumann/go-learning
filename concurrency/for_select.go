package main

import "fmt"

// === FOR SELECT ===
func forSelectPattern() {
	charChannel := make(chan string, 3)
	chars := []string{"a", "b", "c", "d", "e"}

	for _, char := range chars {
		select {
		case charChannel <- char:
			fmt.Println("Sent", char)
		default:
			fmt.Println("Channel is full")
		}
	}

	close(charChannel)

	for char := range charChannel {
		fmt.Println("Received", char)
	}
}
