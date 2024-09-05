package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

/**
 * The program is OOP, which idk is ok to do here.
 * This approach uses an odd/even fork extractment
 * Tweak the rounds and group if you want
 * To see the thinking / eating philosophers in console
 * change printResult = true
 */
var ( // global var
	rounds      = 100000
	group       = 30
	forks       = group
	printResult = false
	waitGroup   sync.WaitGroup
	wg          = &waitGroup
	threadCount = 0
)

type (
	Fork struct {
		id       int
		reserved bool
		mu       *sync.Mutex
	}

	Philosopher struct {
		id                   int
		forkChan1, forkChan2 chan Fork
		eatCount             int
		thinkCount           int
	}
)

func main() {
	if group == 1 { // just to make sure a philosopher wants to eat alone
		forks = 2
	}
	forkArr := make([]chan Fork, forks)
	groupArr := make([]Philosopher, group)

	for i := 0; i < forks; i++ {
		forkArr[i] = make(chan Fork, 1)              // make a channel for the Fork
		forkArr[i] <- Fork{id: i, mu: &sync.Mutex{}} // make a thread for the fork
	}

	for i := 0; i < group; i++ { // birth the philosophers
		groupArr[i] = Philosopher{
			id:        i,
			forkChan1: forkArr[i],
			forkChan2: forkArr[(i+1)%forks], // make sure the last philosopher shares fork with index 0
		}
	}

	for i := 0; i < len(groupArr); i++ {
		wg.Add(1)
		go groupArr[i].run() // start program with the respective threads
	}

	wg.Wait()

	printStats(groupArr) // printStats the results
}

func (p *Philosopher) toString() string {
	return fmt.Sprintf("Philosopher %v ----\n ate: %v times\n though: %v times\n", p.id, p.eatCount, p.thinkCount)
}

func (p *Philosopher) run() {
	defer wg.Done()
	threadCount++

	for i := 0; i < rounds; i++ {
		if rand.Intn(100)%2 == 0 { // 50% they eat
			wg.Add(1)
			if p.id%2 == 0 { // to prevent deadlock, odd/even approach
				go eat(p, p.forkChan1, p.forkChan2)
			} else {
				go eat(p, p.forkChan2, p.forkChan1)
			}
		} else { // 50% they think
			wg.Add(1)
			go p.think()
		}
	}
}

func eat(p *Philosopher, forkChan1 chan Fork, forkChan2 chan Fork) {
	defer wg.Done()
	threadCount++

	f1 := <-forkChan1
	f2 := <-forkChan2

	f1.mu.Lock()
	f2.mu.Lock()

	wg.Add(1)
	go p.eat()

	p.forkChan1 <- f1
	p.forkChan2 <- f2
	f1.mu.Unlock()
	f2.mu.Unlock()
}

func (p *Philosopher) think() { // more threads! more faster!
	defer wg.Done()
	threadCount++

	p.thinkCount++
	if printResult {
		fmt.Println(p.id, "is thinking\n")
	}
	time.Sleep(5000)
}

func (p *Philosopher) eat() { // more threads! more faster!
	defer wg.Done()
	threadCount++
	p.eatCount++
	time.Sleep(5)

	if printResult {
		fmt.Println(p.id, "is eating\n")
	}
}

func printStats(array []Philosopher) {
	fmt.Println("Stats\n")
	for i := 0; i < len(array); i++ {
		fmt.Println(array[i].toString())
	}

	fmt.Println("threadcount:", threadCount)

}
