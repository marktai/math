package pi

import (
	"flag"
	"github.com/cheggaaa/pb"
	"github.com/dustin/go-humanize"
	"log"
	"math"
	"time"
)

// iterates through terms of a partial taylor series
func worker(start, terms int64) float64 {
	cur := start
	var total float64
	negative := false

	if start%2 == 0 {
		return total
	}

	if start%4/2 == 1 {
		negative = true
	}

	for i := int64(0); i < terms; i++ {
		term := float64(1) / float64(cur)

		if negative {
			total -= term
		} else {
			total += term
		}

		cur += 2
		negative = !negative
	}
	return total
}

func SpawnWorkers(terms, start, chunk int64, workers int) float64 {
	cur := start
	workersSpawned := 0
	workersToSpawn := int(math.Ceil(float64(terms) / float64(chunk)))
	outChan := make(chan float64, workers)
	quit := make(chan int, 1)

	if workers > workersToSpawn {
		workers = workersToSpawn
	}

	bar := pb.StartNew(workersToSpawn)

	for i := 0; i < workers; i++ {
		outChan <- float64(0)
	}

	var total float64

	for breaker := false; !breaker; {
		select {
		case result := <-outChan:

			total += result

			go func(start, chunk int64) {
				result := worker(start, chunk)
				bar.Increment()
				outChan <- result
			}(cur, chunk)

			workersSpawned += 1
			cur += chunk * 2

			if workersSpawned >= workersToSpawn {
				quit <- 0
			}
		case <-quit:
			breaker = true
		}
	}

	for i := 0; i < workers; i++ {
		total += <-outChan
	}

	return total * 4
}

func Run() {
	start := flag.Int64("start", 1, "Starting term for pi calculations")
	terms := flag.Int64("terms", 1000000000, "Number of terms to sum")
	exponent := flag.Int("exponent", -1, "10^exponent terms used")
	chunk := flag.Int64("chunk", 10000000, "Number of terms per worker")
	workers := flag.Int("workers", 4, "Number of workers to use")
	flag.Parse()

	realTerms := *terms

	if *exponent != -1 {
		realTerms = int64(math.Pow10(*exponent))
	}

	startTime := time.Now()

	out := SpawnWorkers(realTerms, *start, *chunk, *workers)
	time.Sleep(time.Millisecond * 150)
	log.Println(out)
	log.Println(humanize.Comma(realTerms), "terms took", time.Now().Sub(startTime))
}
