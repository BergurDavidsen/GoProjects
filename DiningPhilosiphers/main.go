package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	rounds        = 10
	group         = 5
	forks         = group
	waitingSeed   = 2
	printActions  = false
	printChannels = true
	printResults  = true
	wg            = &sync.WaitGroup{}
)

type (
	Fork struct {
		id int
	}

	Philosopher struct {
		id                   int
		forkChan1, forkChan2 chan Fork
		eatCount             int
		thinkCount           int
	}
)

func main() {
	if group == 1 {
		forks = 2
	}
	forkArr := make([]chan Fork, forks)
	groupArr := make([]Philosopher, group)

	for i := 0; i < forks; i++ {
		forkArr[i] = make(chan Fork, 1)
		go forkManager(forkArr[i], i)
	}

	for i := 0; i < group; i++ {
		groupArr[i] = Philosopher{
			id:        i,
			forkChan1: forkArr[i],
			forkChan2: forkArr[(i+1)%forks],
		}
	}

	for i := 0; i < len(groupArr); i++ {
		wg.Add(1)
		go groupArr[i].run()
	}

	wg.Wait()
	printStats(groupArr)
}

func forkManager(forkChan chan Fork, id int) {
	for {
		forkChan <- Fork{id: id}
		<-forkChan
	}
}

func (p *Philosopher) run() {
	defer wg.Done()

	for i := 0; i < rounds; i++ {
		if rand.Intn(100)%2 == 0 {
			if p.id%2 == 0 {
				eat(p, p.forkChan1, p.forkChan2)
			} else {
				eat(p, p.forkChan2, p.forkChan1)
			}
		} else {
			think(p)
		}
	}
}

func eat(p *Philosopher, forkChan1 chan Fork, forkChan2 chan Fork) {
	wg.Add(1)
	defer wg.Done()

	// Philosopher acquires forks
	f1 := <-forkChan1
	f2 := <-forkChan2

	if printChannels {
		fmt.Printf("Philosopher: %v, has acquired forks: %v, %v\n", p.id, f1.id, f2.id)
	}

	p.eatCount++
	time.Sleep(randomMillis(waitingSeed))

	if printActions {
		fmt.Println(p.id, "is eating")
	}

	// Philosopher releases forks
	forkChan1 <- f1
	forkChan2 <- f2
	if printChannels {
		fmt.Printf("Philosopher: %v, has released forks: %v, %v\n", p.id, f1.id, f2.id)
	}
}

func think(p *Philosopher) {
	wg.Add(1)
	defer wg.Done()

	p.thinkCount++
	if printActions {
		fmt.Println(p.id, "is thinking")
	}
	time.Sleep(randomMillis(waitingSeed))
}

func randomMillis(seed int) time.Duration {
	return time.Duration(rand.Intn(seed)) * time.Millisecond
}

func (p *Philosopher) toString() string {
	return fmt.Sprintf("Philosopher %v ----\n ate: %v times\n thought: %v times\n", p.id, p.eatCount, p.thinkCount)
}

func printStats(array []Philosopher) {
	fmt.Println("threadcount:", len(array))

	if printResults {
		fmt.Println("STATS:\n")
		for _, p := range array {
			fmt.Println(p.toString())
		}
	}
}
