package main

import (
	"log"
	"time"
	"sync"
	"math/rand"
	"context"
)

func gen(ctx context.Context) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)

		for i := 0; i < 5; i++ {
			if ctx.Err() != nil {
				return
			}

			out <- i
			time.Sleep(time.Duration(rand.Intn(5)) * 500 * time.Millisecond)
		}
	}()

	return out
}

func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func merge(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan int) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

const numWorkers = 20
func main() {
	ctx := context.Background()
	genChans := make([]<-chan int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		genChans[i] = gen(ctx)
	}

	sqChans := make([]<-chan int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		sqChans[i] = sq(merge(genChans...))
	}

	for i := range merge(sqChans...) {
		log.Println(i)
	}
}
