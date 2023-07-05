package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	LABEL_METHOD      = "method"
	LABEL_STATUS_CODE = "status_code"
	LABEL_USER        = "user"
)

var (
	counter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "my_app_http_requests_count",
			Help: "Total amount of HTTP requests",
		},
		[]string{
			LABEL_METHOD,
			LABEL_STATUS_CODE,
			LABEL_USER,
		},
	)

	gauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "my_app_http_requests_in_process",
			Help: "Total amount of HTTP requests being processed at the time",
		},
		[]string{
			LABEL_METHOD,
			LABEL_STATUS_CODE,
			LABEL_USER,
		},
	)

	histogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "my_app_http_requests_latency_seconds",
			Help: "Latency of HTTP requests",
		},
		[]string{
			LABEL_METHOD,
			LABEL_STATUS_CODE,
			LABEL_USER,
		},
	)
)

// HTTP handler
func httpHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	startTime := time.Now()

	// Extract query parameters
	statusCodeParam := r.URL.Query().Get(LABEL_STATUS_CODE)
	user := r.URL.Query().Get(LABEL_USER)

	// Increase in-process requests
	updateInProcessRequests(r.Method, statusCodeParam, user, true)

	// Wait 2 seconds
	time.Sleep(2 * time.Second)

	// Write response
	w.WriteHeader(getCorrespondingStatusCode(statusCodeParam))
	w.Write([]byte("Request is handled."))

	// Decrease in-process requests
	updateInProcessRequests(r.Method, statusCodeParam, user, false)

	// Increment request counter
	incrementRequestCounter(r.Method, statusCodeParam, user)

	// Record request duration
	recordRequestDuration(r.Method, statusCodeParam, user, startTime)
}

func getCorrespondingStatusCode(
	statusCodeParam string,
) int {
	switch statusCodeParam {
	case "200":
		return http.StatusOK
	case "400":
		return http.StatusBadRequest
	case "404":
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func updateInProcessRequests(
	method string,
	statusCode string,
	user string,
	increment bool,
) {
	if increment {
		gauge.With(
			prometheus.Labels{
				LABEL_METHOD:      method,
				LABEL_STATUS_CODE: statusCode,
				LABEL_USER:        user,
			}).Inc()
	} else {
		gauge.With(
			prometheus.Labels{
				LABEL_METHOD:      method,
				LABEL_STATUS_CODE: statusCode,
				LABEL_USER:        user,
			}).Dec()
	}
}

func incrementRequestCounter(
	method string,
	statusCode string,
	user string,
) {
	counter.With(
		prometheus.Labels{
			LABEL_METHOD:      method,
			LABEL_STATUS_CODE: statusCode,
			LABEL_USER:        user,
		}).Inc()
}

func recordRequestDuration(
	method string,
	statusCode string,
	user string,
	startTime time.Time,
) {
	duration := time.Since(startTime).Seconds()
	histogram.With(
		prometheus.Labels{
			LABEL_METHOD:      method,
			LABEL_STATUS_CODE: statusCode,
			LABEL_USER:        user,
		}).Observe(duration)
}
