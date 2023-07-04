package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Start exploring Prometheus!
func main() {

	// Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())

	// App
	http.HandleFunc("/app", httpHandler)

	// Simulate
	go simulate()

	// Serve
	http.ListenAndServe(":8080", nil)
}
