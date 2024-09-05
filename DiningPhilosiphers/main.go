package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	rounds      = 10
	printResult = false
)

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex

	fork0 := make(chan bool, 1)
	fork1 := make(chan bool, 1)
	fork2 := make(chan bool, 1)
	fork3 := make(chan bool, 1)
	fork4 := make(chan bool, 1)

	fork0 <- true
	fork1 <- true
	fork2 <- true
	fork3 <- true
	fork4 <- true

	var p1 = Philosopher{id: 0, forkChan1: fork0, forkChan2: fork1}
	var p2 = Philosopher{id: 1, forkChan1: fork1, forkChan2: fork2}
	var p3 = Philosopher{id: 2, forkChan1: fork2, forkChan2: fork3}
	var p4 = Philosopher{id: 3, forkChan1: fork3, forkChan2: fork4}
	var p5 = Philosopher{id: 4, forkChan1: fork4, forkChan2: fork0}

	wg.Add(5)
	go p1.run(&wg, &mu)
	go p2.run(&wg, &mu)
	go p3.run(&wg, &mu)
	go p4.run(&wg, &mu)
	go p5.run(&wg, &mu)

	wg.Wait()

	fmt.Println("eating count")
	fmt.Println("p1:", p1.eatCount, "\n", "p2:", p2.eatCount, "\n", "p3:", p3.eatCount, "\n", "p4:", p4.eatCount, "\n", "p5", p5.eatCount)

}

type Philosopher struct {
	id                   int
	fork1, fork2         bool
	forkChan1, forkChan2 chan bool
	eatCount             int
	thinkCount           int
}

func (p *Philosopher) run(wg *sync.WaitGroup, mu *sync.Mutex) {
	for i := 0; i < rounds; i++ {
		if rand.Intn(100)%2 == 0 {
			wg.Add(1)
			go eat(p, mu)
			wg.Done()
		} else {
			if printResult {
				wg.Add(1)
				go fmt.Println(p.id, "is thinking")
				wg.Done()
				fmt.Println()
			}
			time.Sleep(5000)
		}
	}

	defer wg.Done()
}

func eat(p *Philosopher, mu *sync.Mutex) {
	mu.Lock()
	<-p.forkChan1
	<-p.forkChan2
	mu.Unlock()

	time.Sleep(5)
	if printResult {
		fmt.Println(p.id, "is eating")
		fmt.Println()
	}
	p.eatCount++

	p.forkChan1 <- true
	p.forkChan2 <- true

}
