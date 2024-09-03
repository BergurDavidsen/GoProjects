package main

import (
	"fmt"
	"math/rand"
)

func main() {
	//minMaxThinking := []int{2, 3, 4, 5, 6, 7}
	forkChannel := make(chan [5]int)

	initForks(forkChannel)
	initPhilosopher(forkChannel)

}
func forks(id int, forksChannel chan [5]int) {
	// Example of sending an ID to the channel
	//<-forksChannel[id]
}

type Philosopher struct {
	id                  int
	leftFork, rightFork int
}

// Corrected function to accept a channel of integers
func philosopher(id int, forksChannel chan [5]int) {
	p := Philosopher{
		id:        id,
		leftFork:  -1,
		rightFork: -1,
	}

	var rightFork int
	var leftFork int

	if ranBool() {
		p.rightFork = requestFork(id)
		p.leftFork = requestFork(id % 4) // size of array 5-1

		if rightFork != -1 || leftFork != -1 {
			isEating(p.id)
		} else {
			isThinking(p.id)
		}
	} else {
		isThinking(id)
	}

}

func requestFork(id int, forkChannel chan []int) int {
	// return -1 if it doesn't work

	return 0
}

func isThinking(id int) {
	fmt.Printf("Philosopher ", id, " is thinking thoughts...")
}

func isEating(id int) {
	fmt.Printf("Philosipher %v is eating pasta", id)
}

func timeLimit() int {
	return rand.Intn(5) + 1
}

func ranBool() bool {
	return rand.Intn(1) == 0
}

// Corrected initPhilosopher function
func initPhilosopher(forks chan [5]int) {
	go philosopher(0, forks)
	go philosopher(1, forks)
	go philosopher(2, forks)
	go philosopher(3, forks)
	go philosopher(4, forks)
}

func initForks(forkChannel chan [5]int) {
	go forks(0, forkChannel)
	go forks(1, forkChannel)
	go forks(2, forkChannel)
	go forks(3, forkChannel)
	go forks(4, forkChannel)
}
