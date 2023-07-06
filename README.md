# Understanding Prometheus with Grafana

This repo is dedicated to show how to generate Prometheus metrics as well as how to scrape and query them per PromQL with Grafana.

## Instrumentation

For demonstration purposes, we will use a Go application. First, get the necessary library which in our case is `github.com/prometheus/client_golang v1.16.0`.

In our [`main.go`](/app/main.go) file, we start the Prometheus exporting as follows:

```golang
// Prometheus metrics
http.Handle("/metrics", promhttp.Handler())
```

Then, we add our HTTP handler:

```golang
// App
http.HandleFunc("/app", httpHandler)
```

which we implemented in [`httphandler.go`](/app/httphandler.go).

We will directly simulate the HTTP handler within the application per [`simulate.go`](/app/simulate.go) and that will run as a separate Go routine:

```golang
// Simulate
go simulate()
```

As last, we will be listening the port `8080` which the simulator will be making random requests to.

```golang
// Serve
http.ListenAndServe(":8080", nil)
```

### Counter metric

The counter is an _always increasing_ type of metric. The moment the application starts, it has the value of zero and the more you increment it, the more it increases and stores it's state. Though, when the application is restarted, the counter is resetted to zero once again.

Therefore, the counter metrics are most useful to track the change of an event, meaning that the derrivate of itself is what matters.

We instatiate a counter metric in our variables as follows:

```golang
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
```

With this metric, we would like to measure the throughput in terms of request per minute (RPM). Therefore, we will increment this counter everytime a request is made to our HTTP handler:

```golang
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
```

### Histogram metric

The histogram is a _bucket of summaries_ type of metric. You define some buckets (which are meaningful for your application) and the histogram metric keeps track of your events with regards to these buckets.

In our case, we instatiate our histogram metric in our variables as follows:

```golang
histogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "my_app_http_requests_latency_seconds",
			Help: "Latency of HTTP requests",
			Buckets: []float64{
				0.00001,
				0.00002,
				0.00005,
				0.00010,
				0.00020,
			},
		},
		[]string{
			LABEL_METHOD,
			LABEL_STATUS_CODE,
			LABEL_USER,
		},
	)
```

Every time a request is made, we will be measuring time spent time and will be recording that duration:

```golang
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
```

### Labels

As you have noticed, we have instantiated our metrics with the following labels:

```golang
const (
	LABEL_METHOD      = "method"
	LABEL_STATUS_CODE = "status_code"
	LABEL_USER        = "user"
)
```

Labels are super important since they provide us with deeper understanding of what our metric is recording.

For every request, we will be tracking the

- HTTP method
- HTTP status code
- User who made the request

This will help us to make an analysis of

- Which user is getting the most `404`s
- Which methods are being used the most
- ...

## Deployment

We will be running 2 instances of our application and will be loading them differently (in sense of RPM).

In the root directory of the repository, you find [`docker-compose.yaml`](/docker-compose.yml). First, we will be needing a network for the containers to talk to each other:

```
networks:
  monitoring:
    driver: bridge
```

Second, we need to store our generated metrics for the Prometheus server somewhere as well as the dashboard we will build for Grafana:

```
volumes:
  prometheus_data: {}
  grafana_data: {}
```

Now, we can define the 2 instances of application under `services`:

```yaml
services:
  # App 1
  app-1:
    container_name: app1
    build:
      context: ./app
      dockerfile: Dockerfile
    environment:
      - REQUEST_INTERVAL=${REQUEST_INTERVAL_APP_1}
    ports:
      - "8080:8080"
    networks:
      - monitoring

  # App 2
  app-2:
    container_name: app2
    build:
      context: ./app
      dockerfile: Dockerfile
    environment:
      - REQUEST_INTERVAL=${REQUEST_INTERVAL_APP_2}
    ports:
      - "8081:8080"
    networks:
      - monitoring
```

As you can see, they accept 2 diffent environment variables which we have defined in [`.env`](/.env) file. This `REQUEST_INTERVAL` variable will be used differently for both instances in order to showcase instance specific filtering and analyzing in Grafana later on.

At last, we have our heros Prometheus and Grafana, respectively:

```yaml
# Prometheus
prometheus:
  image: prom/prometheus:latest
  container_name: prometheus
  restart: unless-stopped
  volumes:
    - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
    - prometheus_data:/prometheus
  command:
    - "--config.file=/etc/prometheus/prometheus.yml"
    - "--storage.tsdb.path=/prometheus"
    - "--web.console.libraries=/etc/prometheus/console_libraries"
    - "--web.console.templates=/etc/prometheus/consoles"
    - "--web.enable-lifecycle"
  ports:
    - "9090:9090"
  networks:
    - monitoring
  depends_on:
    - app-1
    - app-2

# Grafana
grafana:
  image: grafana/grafana:latest
  container_name: grafana
  volumes:
    - grafana_data:/var/lib/grafana
    - ./grafana/provisioning/dashboards:/etc/grafana/provisioning/dashboards
    - ./grafana/provisioning/datasources:/etc/grafana/provisioning/datasources
  environment:
    - GF_SECURITY_ADMIN_USER=${GRAFANA_ADMIN_USERNAME}
    - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
    - GF_USERS_ALLOW_SIGN_UP=false
  restart: unless-stopped
  ports:
    - "3000:3000"
  networks:
    - monitoring
```

### Prometheus

Prometheus needs a configuration file to figure out _where to scrape_. This file is the [prometheus.yml](/prometheus/prometheus.yml):

```yaml
global:
  scrape_interval: 10s

scrape_configs:
  # App 1
  - job_name: "app-1"
    static_configs:
      - targets: ["app-1:8080"]

  # App 2
  - job_name: "app-2"
    static_configs:
      - targets: ["app-2:8080"]
```

We will be scraping the application `/metrics` endpoints every 10 seconds.

### Grafana

Grafana is a visualization tool meaning that it does not store any timeseries for us but just shows them in a proper way. So, it obviously needs some sources to pull the data from which is in our case the Prometheus server that is defined per [`datasource.yml`](/grafana/provisioning/datasources/datasource.yml).

Other than that, I have already built some dashboards which you can refer to out-of-the-box:

- [`count.json`](/grafana/provisioning/dashboards/count.json)
- [`histogram.json`](/grafana/provisioning/dashboards/histogram.json)

## Monitoring

In order to monitor our application, we need to understand the telemetry data we are generating and to do that, we need to retrieve it back from our Prometheus server per PromQL (Prometheus Query Language).

### Getting used to Prometheus

Our Prometheus server exposes a _not-so-friendly_ UI per the port `9090` which we can access through our browser (`http://localhost:9090/`).

Let's execute simply:

```
my_app_http_requests_count
```

which gives us:

```
my_app_http_requests_count{instance="app-1:8080", job="app-1", method="DELETE", status_code="200", user="bill"}
94
my_app_http_requests_count{instance="app-1:8080", job="app-1", method="DELETE", status_code="200", user="elon"}
104
my_app_http_requests_count{instance="app-1:8080", job="app-1", method="DELETE", status_code="200", user="jeff"}
114
my_app_http_requests_count{instance="app-1:8080", job="app-1", method="DELETE", status_code="200", user="warren"}
118
...
```

Well, this stands for all of individual the counter measurements with the given labels and doesn't provide us much. As mentioned before, the `rate` of the counter metrics are important and to do that we will execute the following:

```
rate(my_app_http_requests_count[1m])*60
```

which will calculate the change of increment of the counter in 1 minute buckets and per multiplying it with 60, we get the throughput in terms of RPM that looks like this:

![00_rate_counter.png](/docs/prometheus/00_rate_counter.png)

Since, every unique value of every label corresponds to a complete new timeseries, what we see here is a mess and still doesn't mean much.

Let's sum everything up and try to get the broadest picture:

```
sum (rate(my_app_http_requests_count[1m]))*60
```

![01_rate_counter_sum.png](/docs/prometheus/01_rate_counter_sum.png)

Now, it looks a lot more pure! Let's make a check whether we actually see the correct RPM of the application. The instance 1 of the app is being called every second and the instance 2 every 2 seconds. So, we would expect to see 60 RPM for instance 1 and 30 RPM for instance 2. Since we are summing everything up, in total it should be around 90 RPM and that is exactly what we see in the graph. We're on track!

Now, let's dive a bit deeper. Let's see the RPMs of individual instances:

```
sum by (instance) (rate(my_app_http_requests_count[1m]))*60
```

![02_rate_counter_sum_by_instance.png](/docs/prometheus/02_rate_counter_sum_by_instance.png)

As excepted, the instance 1 has 60 RPM and the instance 2 has 30 RPM.

What if I'm specifically interested in what `elon` has been doing? Let's filter the requests of `elon` and visualize them according to the HTTP status codes:

```
sum by (status_code) (rate(my_app_http_requests_count{user='elon'}[1m]))*60
```

![03_rate_counter_sum_by_status_code_of_elon.png](/docs/prometheus/03_rate_counter_sum_by_status_code_of_elon.png)

### Switching to Grafana

Now, we know how to group and filter our metrics. So, why not use Grafana to create cool dashboards? Grafana can be accessed per `http://localhost:3000` where you can log in with the following super secret credentials:

- username: admin
- password: admin123

Once you log in, you will see 2 pre-built dashboards:

- Counter metric
- Histogram metric

Both of these dashboards are configured with _dashboard variables_ which you will see on top left. By default, all of the `label values` are pre-selected which means that you are seeing the most overall view of your metrics. You can select individual label values to investigate deeper!

**Example:**

This is our `Latency per method (s)` panel in `Histogram Metric` dashboard with all label values selected:

![02_histogram_panel.png](/docs/grafana/02_histogram_panel.png)

and this is the same panel filtered with

- `status_code="200"`
- `instance="app-1:8080"`
- `user=~"jeff,waren"`

![03_histogram_panel_filtered.png](/docs/grafana/03_histogram_panel_filtered.png)

Hope it helped. Enjoy open sourcing!
