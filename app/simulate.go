package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

var (

	// Methods to make random requests
	methods = []string{
		http.MethodGet,
		http.MethodGet,
		http.MethodGet,
		http.MethodGet,
		http.MethodGet,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPost,
		http.MethodPost,
		http.MethodDelete,
		http.MethodDelete,
	}

	// Users to make random requests
	users = []string{
		"elon",
		"warren",
		"jeff",
		"bill",
	}

	// Status codes to be randomely selected
	statusCodes = []string{
		"200",
		"200",
		"200",
		"200",
		"201",
		"201",
		"400",
		"404",
		"500",
	}

	randomizer = rand.New(rand.NewSource(time.Now().UnixNano()))

	// HTTP client
	httpClient = &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
)

// Simulate the own instance of the app
func simulate() {

	// Wait 2 seconds before simulating
	time.Sleep(2 * time.Second)

	for {
		func() {

			// Make request every second
			time.Sleep(time.Second)

			// Prepare HTTP request
			req := prepareHttpRequest()

			// Execute HTTP request
			executeHttpRequest(req)
		}()
	}
}

func prepareHttpRequest() *http.Request {
	user := users[randomizer.Intn(len(users))]
	statusCode := statusCodes[randomizer.Intn(len(statusCodes))]
	method := methods[randomizer.Intn(len(methods))]

	fmt.Println("User: " + user + " | Method: " + method + " | StatusCode: " + statusCode)

	// Create HTTP request with trace context
	req, err := http.NewRequest(
		method,
		"http://localhost:8080/app",
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	// Add headers -> JSON
	req.Header.Add("Content-Type", "application/json")

	// Add request params
	qps := req.URL.Query()
	qps.Add(LABEL_STATUS_CODE, statusCode)
	qps.Add(LABEL_USER, user)
	req.URL.RawQuery = qps.Encode()

	return req
}

func executeHttpRequest(
	req *http.Request,
) {
	// Execute HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	// Read HTTP response
	_, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
}
