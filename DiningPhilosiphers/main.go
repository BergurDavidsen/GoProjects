package main

/**
* Arbi
* Berd
* Broh
* Koad
 */
import (
	"fmt"
	"sync"
)

var (
	wg          = &sync.WaitGroup{}
	request     = make([]chan int, 5)
	release     = make([]chan int, 5)
	philChannel = make([]chan int, 5)
	group       = 5 //needs to be 2 or more
)

func main() {

	// Initialize fork channels with values
	for i := 0; i < 5; i++ {
		// Use buffered channels
		release[i] = make(chan int, 1)
		philChannel[i] = make(chan int, 2)
		request[i] = make(chan int, 1)

	}

	for i := 0; i < group; i++ { // run go routines
		wg.Add(1)
		go philosopher(i)
		go forks(i)
	}
	//wait for done running
	wg.Wait()
}

/**
* The philosopher function
 */
func philosopher(id int) {

	//initialize counters
	eatCount := 0
	thinkCount := 0

	//initialize left and right
	var left int
	var right int

	//To prevent deadlock, we use odd-even approach to make sure some philosopher pick up the same first fork
	if id%2 == 0 {
		left = id
		right = (id + 1) % 5
	} else {
		right = id
		left = (id + 1) % 5
	}

	//main philosopher logic
	for {
		if eatCount >= 3 {
			fmt.Println(id, "is done eating and is very full!😍🥰😘🤢 Thought", thinkCount, "times🤔🧠🧐🔥")

			wg.Done()
			return

		}
		//sends request for forks
		request[left] <- id
		request[right] <- id

		select {
		//picks up first fork if available
		case response := <-philChannel[id]:

			select {
			//picks up second fork if available
			case <-philChannel[id]:
				//has both forks and therfore eats
				fmt.Printf("%d is eating🍽️🍔🍺🍺\n", id)
				eatCount++
				//releases all forks when done eating
				release[left] <- id
				release[right] <- id
			default:
				//second fork not available, relesases first fork
				release[response] <- id
				thinkCount++
				fmt.Printf("%d is thinking🤔💭\n", id)
			}
		default:
			//thinks because no fork is available
			thinkCount++
			fmt.Printf("%d is thinking🤔💭\n", id)

		}
	}

}

/**
* The fork function
 */

func forks(id int) {
	available := true
	for {
		select {
		//checks if philosopher has requested this fork
		case req := <-request[id]:

			if available {
				//give permsission to pick fork up
				available = false
				philChannel[req] <- id
			}
		case <-release[id]:
			//philosopher releases this fork and makes it available for others
			available = true
		}

	}

}
