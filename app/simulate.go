package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func simulate() {

	httpClient := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}

	// Wait 2 seconds before simulating
	time.Sleep(2 * time.Second)

	for {
		func() {

			// Make request every second
			time.Sleep(time.Second)

			// Create HTTP request with trace context
			req, err := http.NewRequest(
				http.MethodGet,
				"http://localhost:8080/app",
				nil,
			)
			if err != nil {
				fmt.Println(err)
			}

			// Add headers
			req.Header.Add("Content-Type", "application/json")

			// Add request params
			qps := req.URL.Query()
			qps.Add("status_code", "200")
			req.URL.RawQuery = qps.Encode()

			// Perform HTTP request

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

			// Check status code
			if res.StatusCode != http.StatusOK {
				fmt.Println(err)
			}
		}()
	}
}
