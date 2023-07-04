package main

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	LABEL_METHOD      = "method"
	LABEL_STATUS_CODE = "status_code"
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
		},
	)
)

// HTTP handler
func httpHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	statusCode := http.StatusOK
	statusCodeParam := r.URL.Query().Get("status_code")

	switch statusCodeParam {
	case "200":
		statusCode = http.StatusOK
	case "400":
		statusCode = http.StatusBadRequest
	case "404":
		statusCode = http.StatusNotFound
	default:
		statusCode = http.StatusInternalServerError
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))

	counter.With(
		prometheus.Labels{
			LABEL_METHOD:      r.Method,
			LABEL_STATUS_CODE: strconv.Itoa(statusCode),
		}).Inc()
}
