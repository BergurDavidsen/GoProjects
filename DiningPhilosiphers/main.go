package main

import (
	"fmt"
	"sync"
)

var (
	wg           sync.WaitGroup
	forksChannel = make([]chan bool, 5)
)

func main() {
	// Initialize fork channels with values
	for i := 0; i < len(forksChannel); i++ {
		forksChannel[i] = make(chan bool, 1) // Use buffered channels
		forksChannel[i] <- true              // Fork is initially available
	}

	for i := 0; i < 5; i++ { // Increase the number of philosophers to match the number of forks
		wg.Add(1)
		go philosopher(i, &wg)
		go forks(i, &wg)
	}
	wg.Wait()
}

func philosopher(id int, wg *sync.WaitGroup) {

	fmt.Println("joe")
	wg.Done()
}

func forks(id int, wg *sync.WaitGroup) {

	fmt.Println("joe")
}
