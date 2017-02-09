package main

import (
	"fmt"
	"net/http"

	"github.com/mfojtik/go-http-metric/pkg/middleware"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}

func main() {
	mid := &middleware.MetricMiddleware{
		Handler:            http.HandlerFunc(myHandler),
		PrometheusEndpoint: "/metric",
	}
	mid.StartMetrics()

	fmt.Println("Listening on port :8080...")
	http.ListenAndServe(":8080", mid)
}
