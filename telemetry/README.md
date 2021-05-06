# Telemetry

[OpenTelemetry](https://opentelemetry.io/docs/go/getting-started/) is being
used to add telemetry to Secretless. For now the scope is to only introduce
metrics for TCP streaming.

The current implementation for metric collection is as follows
1. Define a global meter in the Secretless entrypoint
1. Use Prometheus pull exporter, and expose a metrics endpoint. Ensure that the Prometheus exported has reasonable DefaultHistogramBoundaries. TODO: Find out if they can be set per metric, it seems strange that this is set globally.
1. Create counter for throughput, and recorder for latency. Ensure the metrics are
   labelled with service specific metadata (connector type, service name etc.).
   + `secretless_tcp_stream_throughput`
   + `secretless_tcp_stream_latency`

For making measurements we start by nothing that network I/O in Go is blocking. That
means reads will block a goroutine until the buffer has something. We use
`io.Copy` to implement unidirectional streaming, it takes as input a destination `io.Writer` and
a source `io.Reader`. TCP streaming in Secretless results from 2 Go routines running `io.Copy`
taking as input the client and target TCP connections, each alternating at being source and destination in the 2 Go routines. `io.Copy` handles all the reading,
writing and buffering. The `io.Copy` for each direction blocks in its goroutine
for the lifetime of the streaming.

In order to take latency and throughput measurement, we must
1. Instrument each TCP connection instance by wrapping it to intercept the start and finish
   of reads and writes.
1. Measure latency as the time between when a TCP connection read unblocks, and a write returns. This applies equally to incoming and outgoing streaming, using the client as the datum

## Pending questions

- [ ] What is the impact of Telemetry, if any ?
- [ ] What is a good UX for toggling Telemetry on and off ? 
- [ ] What are the pros and cons of push vs pull metric collection, and how does it impact the data available at analysis time.
- [ ] At present the implementation relies on a Prometheus pull metrics endpoint. What are the configuration options (e.g. polling interval) available and what impact do they have to the data that is available at analysis time.

## Setup

1. Run target service (e.g. postgres on 0.0.0.0:5432)
1. Update [secretless.yml](./secretless.yml) to point to target service
1. Run telemetry infrastructure (Prometheus and Grafana), `docker-compose up -d`
1. Login in to grafana. Credentials are `admin` and `admin`
1. Add prometheus datasource to grafana, URL is `host.docker.internal:9090`
1. Create dashboards using examples in [analysis](#analysis)

## Analysis

Get average latency:
```
rate(secretless_tcp_stream_latency_sum[5m])/rate(secretless_tcp_stream_latency_count[5m])
```

Average latency:
```yaml
Metrics: rate(secretless_tcp_stream_latency_sum[5m])/rate(secretless_tcp_stream_latency_count[5m])
Legend:
```

Latency buckets:
```yaml
Metrics: secretless_tcp_stream_latency_bucket
Legend: {{le}}
```

Latency Histogram:
```yaml
Metrics: histogram_quantile(0.99, sum(rate(secretless_tcp_stream_latency_bucket[5m])) by (le))
Legend: {{le}}
```

Example snapshot from metrics endpoint:
```yaml
➜  ~ curl -X POST -v localhost:2222/metrics
*   Trying 127.0.0.1...
* TCP_NODELAY set
* Connected to localhost (127.0.0.1) port 2222 (#0)
> POST /metrics HTTP/1.1
> Host: localhost:2222
> User-Agent: curl/7.54.0
> Accept: */*
> 
< HTTP/1.1 200 OK
< Content-Type: text/plain; version=0.0.4; charset=utf-8
< Date: Thu, 06 May 2021 11:06:06 GMT
< Transfer-Encoding: chunked
< 
# HELP secretless_tcp_stream_bytes 
# TYPE secretless_tcp_stream_bytes counter
secretless_tcp_stream_bytes{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0"} 383704
# HELP secretless_tcp_stream_latency 
# TYPE secretless_tcp_stream_latency histogram
secretless_tcp_stream_latency_bucket{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0",le="50"} 8043
secretless_tcp_stream_latency_bucket{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0",le="100"} 8706
secretless_tcp_stream_latency_bucket{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0",le="500"} 8794
secretless_tcp_stream_latency_bucket{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0",le="1000"} 8794
secretless_tcp_stream_latency_bucket{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0",le="5000"} 8794
secretless_tcp_stream_latency_bucket{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0",le="10000"} 8794
secretless_tcp_stream_latency_bucket{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0",le="50000"} 8794
secretless_tcp_stream_latency_bucket{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0",le="1e+06"} 8794
secretless_tcp_stream_latency_bucket{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0",le="+Inf"} 8794
secretless_tcp_stream_latency_sum{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0"} 272060
secretless_tcp_stream_latency_count{secretless_connector_name="pg",secretless_service_name="pg_service",service_name="pg:secretless",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="0.20.0"} 8794
* Connection #0 to host localhost left intact
```
