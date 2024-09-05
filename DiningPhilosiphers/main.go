package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Fork struct {
	id       int
	reserved bool
	mu       *sync.Mutex
}

type Philosopher struct {
	id                   int
	forkChan1, forkChan2 chan Fork
	eatCount             int
	thinkCount           int
}

var (
	rounds      = 500000
	printResult = false
	mu          sync.Mutex
)

func main() {
	var wg sync.WaitGroup

	fork0 := make(chan Fork, 1)
	fork1 := make(chan Fork, 1)
	fork2 := make(chan Fork, 1)
	fork3 := make(chan Fork, 1)
	fork4 := make(chan Fork, 1)

	fork0 <- Fork{id: 0, mu: &sync.Mutex{}}
	fork1 <- Fork{id: 1, mu: &sync.Mutex{}}
	fork2 <- Fork{id: 2, mu: &sync.Mutex{}}
	fork3 <- Fork{id: 3, mu: &sync.Mutex{}}
	fork4 <- Fork{id: 4, mu: &sync.Mutex{}}

	var p1 = Philosopher{id: 0, forkChan1: fork0, forkChan2: fork1}
	var p2 = Philosopher{id: 1, forkChan1: fork1, forkChan2: fork2}
	var p3 = Philosopher{id: 2, forkChan1: fork2, forkChan2: fork3}
	var p4 = Philosopher{id: 3, forkChan1: fork3, forkChan2: fork4}
	var p5 = Philosopher{id: 4, forkChan1: fork4, forkChan2: fork0}

	wg.Add(5)
	go p1.run(&wg)
	go p2.run(&wg)
	go p3.run(&wg)
	go p4.run(&wg)
	go p5.run(&wg)

	wg.Wait()
	fmt.Println("p1:", p1.eatCount, "\n", "p2:", p2.eatCount, "\n", "p3:", p3.eatCount, "\n", "p4:", p4.eatCount, "\n", "p5", p5.eatCount)
}

func (p *Philosopher) run(wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < rounds; i++ {
		wg.Add(1) // add work

		if rand.Intn(100)%2 == 0 {
			if p.id%4 == 0 {
				go eat(p, wg, p.forkChan1, p.forkChan2)
			} else {
				go eat(p, wg, p.forkChan2, p.forkChan1)
			}

		} else {
			wg.Done()

			if printResult {
				fmt.Println(p.id, "is thinking\n")
			}
			time.Sleep(5000)
		}
	}
}

func eat(p *Philosopher, wg *sync.WaitGroup, forkChan1 chan Fork, forkChan2 chan Fork) {
	defer wg.Done()

	f1 := <-forkChan1
	f2 := <-forkChan2

	f1.mu.Lock()
	f2.mu.Lock()

	time.Sleep(5)

	if printResult {
		fmt.Println(p.id, "is eating\n")
	}

	p.eatCount++

	p.forkChan1 <- f1
	p.forkChan2 <- f2
	f1.mu.Unlock()
	f2.mu.Unlock()
}
