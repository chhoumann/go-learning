package main

import "fmt"

// === PIPELINE ===

// This is written by Copilot 👇
// A pipeline is a series of stages connected by channels, where each stage is a group of goroutines running the same function.
// In each stage, the goroutines:
// 1. Receive values from upstream via inbound channels
// 2. Perform some function on that data, usually producing new values
// 3. Send values downstream via outbound channels
// Each stage has any number of inbound and outbound channels, except the first and the last stage, which only have outbound and inbound channels, respectively.
// The first stage is sometimes called the source or producer; the last stage, the sink or consumer.
// The diagram below shows a pipeline with three stages, each with one inbound and one outbound channel.
// 	+----------------+   +----------------+   +----------------+
// 	|    Stage 1     |   |    Stage 2     |   |    Stage 3     |
// 	|                |   |                |   |                |
// 	| in  ->  [ ]  ->|-> | in  ->  [ ]  ->|-> | in  ->  [ ]  ->|
// 	|                |   |                |   |                |
// 	|                |   |                |   |                |
// 	| out <-  [ ] <- |<- | out <-  [ ] <- |<- | out <-  [ ] <- |
// 	+----------------+   +----------------+   +----------------+
// The diagram shows the data flow in the pipeline. The arrows represent the direction of the data flow. The boxes represent the goroutines that make up the pipeline stages. The channels are not shown in the diagram, but they connect the goroutines.
// The pipeline pattern is a natural way to express concurrent computation. It can be used to parallelize CPU-bound work, I/O-bound work, or to distribute work across a cluster.
// The pipeline pattern can be used to express a wide variety of concurrent computations. For example, it can be used to:
// 1. Read data from a file, process it, and write it to another file
// 2. Download web pages in parallel, parse them, and do something with the data
// 3. Compress files concurrently
// 4. Generate and solve instances of a problem in parallel
// 5. Distribute work across a cluster
// 6. And many other things
// The pipeline pattern is a powerful way to express concurrent computation. It is simple, flexible, and composable. It can be used to express a wide variety of concurrent computations, from simple to complex.

// sliceToChannel converts a slice of integers to a channel of integers.
// It starts a goroutine to send the integers on the channel and closes the channel when all the integers have been sent.
func sliceToChannel(nums []int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for _, num := range nums {
			fmt.Println("SliceToChannel | Sent", num)
			out <- num
		}
	}()

	return out
}

// square reads integers from in, squares them, and sends the results on a new channel.
// It closes the new channel when in is closed.
func square(in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for num := range in {
			fmt.Println("Square | Received", num)
			out <- num * num
		}
	}()

	return out
}

// sum reads integers from in, sums them, and sends the result on a new channel.
// It closes the new channel when in is closed.
func sum(in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		sum := 0
		for num := range in {
			fmt.Println("Sum | Received", num)
			sum += num
		}
		out <- sum
	}()

	return out
}

func pipelinePattern() {
	nums := []int{7, 2, 8, -9, 4, 0}

	// stage 1
	dataChannel := sliceToChannel(nums)
	// stage 2
	squaredChannel := square(dataChannel)
	// stage 3
	resultChannel := sum(squaredChannel)

	for result := range resultChannel {
		fmt.Println("Result: ", result)
	}
}