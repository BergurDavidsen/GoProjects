package main

import (
	"fmt"
	"sync"
)

var (
	wg           = &sync.WaitGroup{}
	forksChannel = make([]chan int, 5)
	philChannel  = make([]chan int, 5)
)

func main() {
	// Initialize fork channels with values
	for i := 0; i < len(forksChannel); i++ {
		forksChannel[i] = make(chan int, 1) // Use buffered channels
		philChannel[i] = make(chan int, 1)
		forksChannel[i] <- i
	}

	for i := 0; i < 5; i++ { // Increase the number of philosophers to match the number of forks
		wg.Add(1)
		go philosopher(i, wg)
		go forks(i, wg)
	}
	wg.Wait()
}

func philosopher(id int, wg *sync.WaitGroup) {
	left := id
	right := (id + 1) % 5
	forksChannel[left] <- left
	forksChannel[right] <- right

	for i := 0; i < 5; i++ {

		select {
		case <-forksChannel[left]:
			select {
			case <-forksChannel[right]:
				fmt.Printf("%d is eating", id)
			default:
				fmt.Printf("%d is thinking", id)
			}
		}
	}

	wg.Done()
}

func forks(id int, wg *sync.WaitGroup) {
	<-forksChannel[id]

}
