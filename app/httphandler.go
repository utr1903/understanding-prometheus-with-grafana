package main

import (
	"net/http"

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
)

// HTTP handler
func httpHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	statusCodeParam := r.URL.Query().Get(LABEL_STATUS_CODE)
	user := r.URL.Query().Get(LABEL_USER)

	w.WriteHeader(getCorrespondingStatusCode(statusCodeParam))
	w.Write([]byte("Request is handled."))

	counter.With(
		prometheus.Labels{
			LABEL_METHOD:      r.Method,
			LABEL_STATUS_CODE: statusCodeParam,
			LABEL_USER:        user,
		}).Inc()
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
