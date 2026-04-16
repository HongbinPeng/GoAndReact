package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Product struct {
	Name string
}

func Producer(id int, jobs <-chan int, results chan<- Product, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		fmt.Println("Producer", id, "producing", job)
		time.Sleep(1 * time.Second) // simulate producing
		product := Product{Name: "Product" + strconv.Itoa(job) + "-Byproducer" + strconv.Itoa(id)}
		results <- product
	}
}

func Consumer(id int, results <-chan Product, wg *sync.WaitGroup) {
	defer wg.Done()

	for product := range results {
		fmt.Println("Consumer", id, "consuming", product.Name)
		time.Sleep(2 * time.Second) // simulate consuming
	}
}

func main() {
	jobs := make(chan int, 100)
	results := make(chan Product, 100)
	const workerCount = 5
	var producerWG sync.WaitGroup
	var consumerWG sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		producerWG.Add(1)
		go Producer(i, jobs, results, &producerWG)

		consumerWG.Add(1)
		go Consumer(i, results, &consumerWG)
	}
	for i := 0; i < 10; i++ {
		jobs <- i
	}
	close(jobs)
	// Close results only after every producer has finished sending.
	producerWG.Wait()
	close(results)
	// Wait until every consumer has drained the results channel.
	consumerWG.Wait()
}
