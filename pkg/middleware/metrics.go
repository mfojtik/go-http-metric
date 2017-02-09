package middleware

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricMiddleware is a go-http middleware add metrics reporting for the http.Handler.
// The metrics include standard Go runtime metrics but also total number of requests
// served by this HTTP server as well as requests per second counter.
type MetricMiddleware struct {
	// Handler represents the user HTTP handler
	Handler http.Handler
	// PrometheusEndpoint is an HTTP endpointn for prometheus
	PrometheusEndpoint string
	// RequestPerSecondFunc is a function that runs with updated request/per/second
	RequestPerSecondFunc func(int64)
	// Name is an identifier used in prometheus
	Name string

	metricSink metrics.MetricSink

	counterLock *sync.Mutex
	counter     int64

	counterChan chan struct{}
	resetChan   chan struct{}

	requestsPerSecond chan int64
}

func (a *MetricMiddleware) StartMetrics() {
	a.initialize()
}

func (a *MetricMiddleware) initialize() {
	if a.counterChan != nil {
		return
	}
	var err error
	if a.metricSink, err = prometheus.NewPrometheusSink(); err != nil {
		log.Fatalf("metrics: error occured during prometheus sink initialization: %v", err)
	}
	if len(a.Name) == 0 {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatalf("metrics: unable to determine the hostname, please set name manually: %v", err)
		}
		a.Name = hostname
	}
	a.counterChan = make(chan struct{})
	a.resetChan = make(chan struct{})
	a.requestsPerSecond = make(chan int64, 1)
	a.counterLock = &sync.Mutex{}
	go a.watchRequestsPerSecond()
	go a.watchCounter()
	go a.watchReset()
	go a.startReset()
}

func (a *MetricMiddleware) watchRequestsPerSecond() {
	for {
		select {
		case current, ok := <-a.requestsPerSecond:
			if !ok {
				return
			}
			a.metricSink.AddSample([]string{a.Name, "requestPerSecond"}, float32(current))
			a.metricSink.SetGauge([]string{a.Name, "requestPerSecondCurrent"}, float32(current))
			if a.RequestPerSecondFunc != nil {
				a.RequestPerSecondFunc(current)
			}
		}
	}
}

// startReset defines a one second timer that resets the counter.
func (a *MetricMiddleware) startReset() {
	for {
		select {
		case <-time.After(1 * time.Second):
			a.reset()
		}
	}
}

// watchReset observes the reset requests and records the current requests per seconds
// before it resets the counter.
func (a *MetricMiddleware) watchReset() {
	for {
		select {
		case _, ok := <-a.resetChan:
			if !ok {
				return
			}
			a.counterLock.Lock()
			a.requestsPerSecond <- a.counter
			a.counter = 0
			a.counterLock.Unlock()
		}
	}
}

// watchCounter observes the counter increment requests and increments the counter.
func (a *MetricMiddleware) watchCounter() {
	for {
		select {
		case _, ok := <-a.counterChan:
			if !ok {
				a.cleanup()
				return
			}
			a.counterLock.Lock()
			a.counter += 1
			a.metricSink.IncrCounter([]string{a.Name, "totalRequests"}, 1)
			a.counterLock.Unlock()
		}
	}
}

func (a *MetricMiddleware) cleanup() {
	close(a.resetChan)
	close(a.requestsPerSecond)
}

func (a *MetricMiddleware) reset() {
	a.resetChan <- struct{}{}
}

func (a *MetricMiddleware) increment() {
	a.counterChan <- struct{}{}
}

// ServeHTTP wraps the go-http ServerHTTP method with metrics.
func (a *MetricMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.initialize()
	if len(a.PrometheusEndpoint) > 0 && r.RequestURI == a.PrometheusEndpoint {
		promhttp.Handler().ServeHTTP(w, r)
		return
	}
	start := time.Now()
	defer a.metricSink.SetGauge([]string{a.Name, "requestTimeNanoseconds"}, float32(time.Since(start).Nanoseconds()))
	a.increment()
	a.Handler.ServeHTTP(w, r)
}
