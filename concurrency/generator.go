package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

// === GENERATOR ===
// repeatFunc repeats the result of fn until done is closed.
// It sends the results on a new channel and closes the channel when done is closed.
func repeatFunc[T any, K any](done <-chan K, fn func() T) <-chan T {
	stream := make(chan T)

	go func() {
		defer close(stream)

		for {
			select {
			case <-done:
				return
			case stream <- fn():
			}
		}
	}()

	return stream
}

// this allows us to take n values from in channel and write to taken channel
func take[T any, K any](done <-chan K, in <-chan T, n int) <-chan T {
	taken := make(chan T)

	go func() {
		defer close(taken)

		// for-select pattern
		// take n values from in channel and write to taken channel
		for i := 0; i < n; i++ {
			select {
			case <-done:
				return
			// read from in channel and write to taken channel
			case taken <- <-in:
			}
		}
	}()

	return taken
}

func primeFinder(done <-chan bool, intStream <-chan int) <-chan int {
	isPrime := func(n int) bool {
		if n <= 1 {
			return false
		}
		for i := 2; i < n; i++ {
			if n%i == 0 {
				return false
			}
		}
		return true
	}

	primeStream := make(chan int)

	go func() {
		defer close(primeStream)
		for {
			select {
			case <-done:
				return
			case num := <-intStream:
				if isPrime(num) {
					primeStream <- num
				}
			}
		}
	}()

	return primeStream
}

func fanIn[T any](done <-chan bool, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	fannedInStream := make(chan T)

	// fan in
	transfer := func(c <-chan T) {
		defer wg.Done()
		for val := range c {
			select {
			case <-done:
				return
			case fannedInStream <- val:
			}
		}
	}

	// could do `wg.Add(len(channels))` here instead of add 1 inside
	for _, c := range channels {
		wg.Add(1) 
		go transfer(c)
	}

	go func() {
		wg.Wait()
		close(fannedInStream)
	}()

	return fannedInStream
}

// showcase generator pattern (+ pipeline) with primeFinder and fanning out & in
// fanning out means to split the work into multiple goroutines
// fanning in means to combine the results from multiple goroutines
func generatorPattern() {
	start := time.Now()
	done := make(chan bool)
	defer close(done)

	randNumFetcher := func() int { return rand.Intn(100) }
	randIntStream := repeatFunc(done, randNumFetcher)

	// naive & slow
	// primeStream := primeFinder(done, randIntStream)
	// for random := range take(done, primeStream, 10) {
	// 	fmt.Println(random)
	// }

	// fan out
	CPUCount := runtime.NumCPU() // spin up as many goroutines as there are CPUs
	primeStreams := make([]<-chan int, CPUCount) // to simulate some slow computation

	for i := 0; i < CPUCount; i++ {
		primeStreams[i] = primeFinder(done, randIntStream)
	}

	// fan in
	fannedInStream := fanIn(done, primeStreams...)

	for randomPrime := range take(done, fannedInStream, 10) {
		fmt.Println(randomPrime)
	}

	fmt.Println("Time taken: ", time.Since(start))
}