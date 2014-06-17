package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	clients            int
	lambda             float64
	timeout            time.Duration
	exponentialBackoff bool
	simpleBackoffDelay time.Duration
)

const (
	server              = "http://127.0.0.1:8070/api"
	expMinDelay         = 100 * time.Millisecond
	expMaxDelay         = 5 * time.Minute
	expFactor   float64 = 2.71828
	expJitter   float64 = 0.1
)

type Counter struct {
	values   [2]uint32
	active   int
	interval time.Duration
}

func NewCounter(interval time.Duration) *Counter {

	c := &Counter{interval: interval}

	go func() {
		timer := time.Tick(c.interval)
		for _ = range timer {
			atomic.StoreUint32(&c.values[1-c.active], 0)
			c.active = 1 - c.active
		}
	}()

	return c
}

func (c *Counter) Increment() {
	atomic.AddUint32(&c.values[c.active], 1)
}

func (c *Counter) Value() float64 {
	return float64(c.values[1-c.active]) / float64(c.interval/time.Second)
}

var (
	counterOk, counterTimeout, counterError *Counter
)

func exponentialDistribution(lambda float64) float64 {
	return -math.Log(1.0-rand.Float64()) * lambda
}

func client() {
	var (
		lastError = false
		lastDelay time.Duration
	)

	for {
		var delay time.Duration

		if lastError {
			if exponentialBackoff {
				delay = time.Duration(float64(lastDelay) * expFactor)
				if delay > expMaxDelay {
					delay = expMaxDelay
				}
				delay += time.Duration(rand.NormFloat64() * expJitter * float64(time.Second))
			} else {
				delay = simpleBackoffDelay
			}
		} else {
			delay = time.Duration(exponentialDistribution(lambda) * float64(time.Second))
		}
		lastDelay = delay
		<-time.After(delay)

		ch := make(chan error)
		req, _ := http.NewRequest("GET", server, nil)

		go func() {
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				resp.Body.Close()
			}

			ch <- err
		}()

		select {
		case err := <-ch:
			if err == nil {
				counterOk.Increment()
				lastError = false
				lastDelay = expMinDelay
			} else {
				counterError.Increment()
				lastError = true
			}
		case <-time.After(timeout):
			http.DefaultTransport.(*http.Transport).CancelRequest(req)
			counterTimeout.Increment()
			lastError = true
		}
	}
}

func main() {
	flag.IntVar(&clients, "clients", 1000, "number of clients")
	flag.Float64Var(&lambda, "lambda", 10.0, "lambda of distribution")
	flag.DurationVar(&timeout, "timeout", 2*time.Second, "timeout on HTTP request")
	flag.BoolVar(&exponentialBackoff, "exponential-backoff", false, "enable exponential backoff")
	flag.DurationVar(&simpleBackoffDelay, "simple-backoff-delay", 100*time.Millisecond, "delay when doing simple backoff")
	flag.Parse()

	const interval = 5 * time.Second

	counterOk = NewCounter(interval)
	counterError = NewCounter(interval)
	counterTimeout = NewCounter(interval)

	for i := 0; i < clients; i++ {
		go client()
	}

	for _ = range time.Tick(interval) {
		fmt.Printf("OK: %.2f req/sec, errors: %.2f req/sec, timedout: %.2f req/sec\n",
			counterOk.Value(), counterError.Value(), counterTimeout.Value())
	}
}
