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
 * The print booleans are if you want them printed
 * in the console
 */
var ( // global var
	rounds        = 10
	group         = 5
	forks         = group
	waitingSeed   = 20000000 // random ms from 0-n
	printActions  = !true
	printChannels = !false
	printResults  = !true
	threadCount   = 0
	wg            = &sync.WaitGroup{}
)

type (
	Fork struct {
		id       int
		reserved bool
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
		wg.Add(1)
		forkArr[i] = make(chan Fork, 1) // make a channel for the Fork
		f := Fork{id: i}
		forkArr[i] <- f // insert fork into channel
		go f.run()      // Start a go routine
	}

	for i := 0; i < group; i++ { // birth the philosophers
		wg.Add(1)

		groupArr[i] = Philosopher{
			id:        i,
			forkChan1: forkArr[i],
			forkChan2: forkArr[(i+1)%forks], // make sure the last philosopher shares fork with index 0
		}

		go groupArr[i].run() // start program with the respective threads

	}

	wg.Wait()

	printStats(groupArr)
}

func (p *Philosopher) run() {
	defer wg.Done()
	threadCount++

	for i := 0; i < rounds; i++ {
		if rand.Intn(100)%2 == 0 { // 50% they eat
			if p.id%2 == 0 { // to prevent deadlock, odd/even approach
				eat(p, p.forkChan1, p.forkChan2)
			} else {
				eat(p, p.forkChan2, p.forkChan1)
			}
		} else { // 50% they think
			think(p)
		}
	}
}

func eat(p *Philosopher, forkChan1 chan Fork, forkChan2 chan Fork) {
	threadCount++

	f1 := <-forkChan1 // fc 2
	f2 := <-forkChan2 // fc 1

	if printChannels {
		fmt.Printf("Philosopher: %v, has aquired forks: %v, %v\n", p.id, f1.id, f2.id)
	}

	p.eatCount++
	time.Sleep(randomMillis(waitingSeed))

	if printActions {
		fmt.Println(p.id, "is eating\n")
	}

	forkChan1 <- f1 // f2
	forkChan2 <- f2 // f1
	if printChannels {
		fmt.Printf("Philosopher: %v, has released forks: %v, %v\n", p.id, f1.id, f2.id)
	}

}

func (f *Fork) run() {
	threadCount++
	wg.Done()
	wg.Wait()
}

func think(p *Philosopher) {
	p.thinkCount++
	if printActions {
		fmt.Println(p.id, "is thinking\n")
	}
	time.Sleep(randomMillis(waitingSeed))
}

func randomMillis(seed int) time.Duration {
	return time.Duration(rand.Intn(seed))
}

func (p *Philosopher) toString() string {
	return fmt.Sprintf("Philosopher %v ----\n ate: %v times\n though: %v times\n", p.id, p.eatCount, p.thinkCount)
}

func printStats(array []Philosopher) {
	fmt.Println("threadcount:", threadCount)

	if printResults {
		fmt.Println("Stats\n")
		for i := 0; i < len(array); i++ {
			fmt.Println(array[i].toString())
		}
	}

}
