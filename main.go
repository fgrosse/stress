package main

import (
	"flag"
	"runtime"
	"sync"
)

func main() {
	numWorkers := flag.Int("workers", runtime.NumCPU(), "Number of workers to run")
	flag.Parse()

	runWorkers(*numWorkers)
}

func runWorkers(n int) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(&wg)
	}

	wg.Wait() // Wait for all workers (in this case, indefinitely)
}

func worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		fibonacci(35) // Adjust 35 to control workload intensity
	}
}

// Fibonacci function without recursion, using iteration instead.
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	prev, curr := 0, 1
	for i := 2; i <= n; i++ {
		prev, curr = curr, prev+curr
	}
	return curr
}
