package main

import (
	"fmt"
	"math/big"
	"time"
)

func main() {
	// Define the target number as a big.Int
	target := new(big.Int)
	target.SetString("100000000", 10) // Base 10 representation

	// Create a counter initialized to 0
	counter := new(big.Int)
	step := new(big.Int).SetInt64(1)                // Increment step
	progressInterval := new(big.Int).SetInt64(1e18) // Progress print interval (10^18)

	// Start the timer
	startTime := time.Now()

	// Loop until the counter reaches the target
	for counter.Cmp(target) < 0 { // Compare counter < target
		counter.Add(counter, step) // Increment the counter

		// Print progress and speed at defined intervals
		if counter.Mod(counter, progressInterval).Cmp(big.NewInt(0)) == 0 {
			// Calculate elapsed time and speed
			elapsed := time.Since(startTime).Seconds()
			speed := new(big.Int).Div(counter, big.NewInt(int64(elapsed))) // Counts per second
			fmt.Printf("Current count: %s | Speed: %s counts/sec | Time elapsed: %.2f seconds\n",
				counter.String(), speed.String(), elapsed)
		}
	}

	// Print final message
	totalElapsed := time.Since(startTime).Seconds()
	fmt.Printf("Counting complete! Reached: %s in %.2f seconds.\n", counter.String(), totalElapsed)
}
