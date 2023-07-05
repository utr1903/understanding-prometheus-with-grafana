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
)

// HTTP handler
func httpHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	// Extract query parameters
	statusCodeParam := r.URL.Query().Get(LABEL_STATUS_CODE)
	user := r.URL.Query().Get(LABEL_USER)

	// Increase in-process requests
	updateInProcessRequests(true, r.Method, statusCodeParam, user)

	// Wait 2 seconds
	time.Sleep(2 * time.Second)

	// Write response
	w.WriteHeader(getCorrespondingStatusCode(statusCodeParam))
	w.Write([]byte("Request is handled."))

	// Decrease in-process requests
	updateInProcessRequests(false, r.Method, statusCodeParam, user)

	// Increment request counter
	incrementRequestCounter(r.Method, statusCodeParam, user)
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
	increment bool,
	method string,
	statusCode string,
	user string,
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
