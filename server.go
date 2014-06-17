package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	currentConcurrency int32
	concurrencyLimit   int
	factor             float64
	baseDelay          time.Duration
	lastDelay          time.Duration
)

func getDelay() time.Duration {
	delay := baseDelay
	current := int(currentConcurrency)

	if current > concurrencyLimit {
		delay = time.Duration(float64(delay) * math.Pow(factor, float64(current-concurrencyLimit)/15.0))
	}

	return delay
}

func httpHandler(rw http.ResponseWriter, req *http.Request) {
	atomic.AddInt32(&currentConcurrency, 1)

	start := time.Now()

	for {
		<-time.After(50 * time.Millisecond)
		delay := getDelay()
		lastDelay = delay

		elapsed := time.Now().Sub(start)

		if elapsed > delay {
			break
		}
	}

	rw.Header().Add("Context-Type", "text/text")
	rw.Write([]byte("OK"))

	atomic.AddInt32(&currentConcurrency, -1)
}

func reporter() {
	for _ = range time.Tick(time.Second) {
		fmt.Printf("%s: concurrency: %4d, last delay: %s\n", time.Now().Format(time.StampMilli), currentConcurrency, lastDelay.String())
	}
}

func main() {
	flag.IntVar(&concurrencyLimit, "limit", 30, "limit on concurrency")
	flag.Float64Var(&factor, "factor", 1.05, "factor of increasing delay")
	flag.DurationVar(&baseDelay, "baseDelay", 100*time.Millisecond, "base service delay")
	flag.Parse()

	go reporter()
	http.HandleFunc("/api", httpHandler)
	log.Fatal(http.ListenAndServe(":8070", nil))
}
