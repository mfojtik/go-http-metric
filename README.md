go-http-metric
==========

Go [net/http](https://golang.org/pkg/net/http/) middleware that provides various metrics
about the HTTP server via the Prometheus endpoint.

Additionally a custom `RequestPerSecond` function can be registred that receives the
current amount of requests per second the server is serving to clients.

This can be used to either collect metrics into Prometheus or provide a base for
auto-scaling applications.

Usage
-----

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/mfojtik/go-http-metric/pkg/middleware"
)

// myHandler is a custom HTTP handler
func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}

func main() {
	mid := &middleware.MetricMiddleware{
    // Register the custom HTTP handler
		Handler:            http.HandlerFunc(myHandler),
    // Prometheus endpoint (optional)
		PrometheusEndpoint: "/metric",
	}
  // Start collected metrics (optional, initialization takes place automatically
  // when a first request comes in)
	mid.StartMetrics()

	fmt.Println("Listening on port :8080...")
	http.ListenAndServe(":8080", mid)
}
```

Optionally you can set `Name` field to give the server identity (all prometheus metric
fields will be prefixed with this name). By default the name is set to a hostname.

Running this example will provide following when you `GET /metric`:

```
...
...
# HELP sigma_local_requestPerSecond sigma_local_requestPerSecond
# TYPE sigma_local_requestPerSecond summary
sigma_local_requestPerSecond{quantile="0.5"} 0
sigma_local_requestPerSecond{quantile="0.9"} 2
sigma_local_requestPerSecond{quantile="0.99"} 2
sigma_local_requestPerSecond_sum 219
sigma_local_requestPerSecond_count 43
# HELP sigma_local_requestPerSecondCurrent sigma_local_requestPerSecondCurrent
# TYPE sigma_local_requestPerSecondCurrent gauge
sigma_local_requestPerSecondCurrent 0
# HELP sigma_local_requestTimeNanoseconds sigma_local_requestTimeNanoseconds
# TYPE sigma_local_requestTimeNanoseconds gauge
sigma_local_requestTimeNanoseconds 135
# HELP sigma_local_totalRequests sigma_local_totalRequests
# TYPE sigma_local_totalRequests counter
sigma_local_totalRequests 219
```

Installation
------------

```sh
go get github.com/mfojtik/go-http-metric
```

Get Go dependencies:
```sh
cd $GOPATH/src/github.com/mfojtik/go-http-metric && go get -a ./pkg/...
```

Todo
-----

[ ] Add unit tests
