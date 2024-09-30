package main

import (
	"fmt"
	"sync"
)

/**
 * André <arbi@itu.dk>
 * Bergur <berd@itu.dk>
 * Bror <broh@itu.dk>
 * Konrad <koad@itu.dk>
 *
 * The reason this program doesn't deadlock are depicted on l:37 and l:73
 *
 * Explanation
 * l:37 The alternating assignment of left and right forks prevents circular wait
 * l:73 The use of channels to request and release forks, coupled with this odd-even strategy,
 * ensures that resources (forks) are available in a way that avoids deadlock.
 */

// the size of the dining philosopher group,
// and some general channels in which the fork and philosopher communicates through
var (
	wg          = &sync.WaitGroup{}
	group       = 5 // must be > 1; 1 philosopher must not eat alone
	request     = make([]chan int, group)
	release     = make([]chan int, group)
	philChannel = make([]chan int, group)
)

/**
 * philosopher simulates a philosopher who alternates between thinking and eating.
 * Each philosopher can only eat if they acquire both forks. To avoid deadlock,
 * philosophers use an odd-even approach when picking up forks.
 *
 * @param id The unique identifier for the philosopher (0 to group-1)
 */
func philosopher(id int) {
	// initialize counters
	eatCount := 0
	thinkCount := 0

	// initialize left and right id for their fork
	var left int
	var right int

	// to prevent deadlock, we use odd-even approach.
	if id%2 == 0 {
		left = id
		right = (id + 1) % group
	} else {
		right = id
		left = (id + 1) % group
	}

	// main philosopher logic
	for {
		if eatCount >= 3 {
			fmt.Println(id, "is done eating and is very full!😍🥰😘🤢 Thought", thinkCount, "times🤔🧠🧐🔥")
			wg.Done()
			return
		}

		// sends request for forks
		request[left] <- id
		request[right] <- id

		select {
		// picks up first fork if available
		case response := <-philChannel[id]:

			select {
			// picks up second fork if available
			case <-philChannel[id]:
				// has both forks and therefore eats
				fmt.Printf("%d is eating🍽️🍔🍺🍺\n", id)
				eatCount++
				// releases all forks when done eating
				release[left] <- id
				release[right] <- id
			default:
				// second fork not available, releases first fork
				release[response] <- id
				thinkCount++
				fmt.Printf("%d is thinking🤔💭\n", id)
			}
		default:
			// thinks because no fork is available
			thinkCount++
			fmt.Printf("%d is thinking🤔💭\n", id)

		}
	}
}

/**
 * fork manages the availability of a single fork in the Dining Philosophers problem.
 * It listens for requests from philosophers to pick up or release the fork.
 */
func fork(id int) {
	available := true
	for {
		select {
		// checks if a philosopher has requested this fork
		case req := <-request[id]:

			if available {
				// give permission to pick fork up
				available = false
				philChannel[req] <- id
			}
		case <-release[id]:
			// philosopher releases this fork and makes it available for others
			available = true
		}
	}
}

/**
 * main initializes the channels and starts the simulation of philosophers and forks.
 * It launches goroutines for each philosopher and fork, and waits for the simulation to finish.
 */
func main() {
	// initialize fork channels with values
	for i := 0; i < group; i++ {
		// use buffered channels
		release[i] = make(chan int, 1)
		philChannel[i] = make(chan int, 2)
		request[i] = make(chan int, 1)

	}

	for i := 0; i < group; i++ { // run go routines
		wg.Add(1)
		go philosopher(i)
		go fork(i)
	}
	// wait for done running
	wg.Wait()
}
