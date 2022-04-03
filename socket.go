package main

import (
	"io"
	"net"
	"os"
	"syscall"

	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

type upstream struct {
	Latency        float64 `json:"upstreamLatency"`
	ResponseLength float64 `json:"upstreamResponseLength"`
	ResponseTime   float64 `json:"upstreamResponseTime"`
	// Status         string  `json:"upstreamStatus"`
	CacheStatus    string  `json:"upstreamCacheStatus"`
}

type socketData struct {
	Status         string  `json:"status"`
	ResponseLength float64 `json:"responseLength"`
	Method         string  `json:"method"`
	RequestLength  float64 `json:"requestLength"`
	RequestTime    float64 `json:"requestTime"`
	upstream
	Service string `json:"service"`
	Path    string `json:"path"`
}

// SocketCollector stores prometheus metrics and ingress meta-data
type SocketCollector struct {
	prometheus.Collector
	requestTime     *prometheus.HistogramVec
	requestLength   *prometheus.HistogramVec
	responseTime    *prometheus.HistogramVec
	responseLength  *prometheus.HistogramVec
	upstreamLatency *prometheus.SummaryVec
	bytesSent       *prometheus.HistogramVec
	requests        *prometheus.CounterVec
	listener        net.Listener
	metricMapping   map[string]interface{}
}

var (
	requestTags = []string{
		"status",
		"method",
		"path",
		"service",
		// "upstream_status",
		"upstream_cache_status",
	}
)

// DefObjectives was removed in https://github.com/prometheus/client_golang/pull/262
// updating the library to latest version changed the output of the metrics
var defObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}

// NewSocketCollector creates a new SocketCollector instance using
// the ingress watch namespace and class used by the controller
func NewSocketCollector() (*SocketCollector, error) {
	socket := "/tmp/prometheus-nginx.socket"
	// unix sockets must be unlink()ed before being used
	_ = syscall.Unlink(socket)

	listener, err := net.Listen("unix", socket)
	if err != nil {
		return nil, err
	}

	err = os.Chmod(socket, 0777) // #nosec
	if err != nil {
		return nil, err
	}

	klog.InfoS("socket is ready /tmp/prometheus-nginx.socket")

	constLabels := prometheus.Labels{
		// "controller_namespace": namespace,
		// "controller_class":     class,
		// "controller_pod":       pod,
	}

	requestTags := requestTags

	sc := &SocketCollector{
		listener: listener,
		responseTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "response_duration_seconds",
				Help:        "The time spent on receiving the response from the upstream server",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
			},
			requestTags,
		),
		responseLength: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "response_size",
				Help:        "The response length (including request line, header, and request body)",
				Namespace:   PrometheusNamespace,
				Buckets:     prometheus.ExponentialBuckets(2, 4, 14),
				ConstLabels: constLabels,
			},
			requestTags,
		),

		requestTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "request_duration_seconds",
				Help:        "The request processing time in milliseconds",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
			},
			requestTags,
		),
		requestLength: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "request_size",
				Help:        "The request length (including request line, header, and request body)",
				Namespace:   PrometheusNamespace,
				Buckets:     prometheus.ExponentialBuckets(32, 2, 12),
				ConstLabels: constLabels,
			},
			requestTags,
		),

		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "requests",
				Help:        "The total number of client requests.",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
			},
			[]string{"status", "service"},
		),

		bytesSent: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "bytes_sent",
				Help:        "The number of bytes sent to a client",
				Namespace:   PrometheusNamespace,
				Buckets:     prometheus.ExponentialBuckets(2, 4, 14),
				ConstLabels: constLabels,
			},
			requestTags,
		),

		upstreamLatency: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:        "upstream_latency_seconds",
				Help:        "Upstream service latency per service",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
				Objectives:  defObjectives,
			},
			[]string{"service"},
		),
	}

	sc.metricMapping = map[string]interface{}{
		prometheus.BuildFQName(PrometheusNamespace, "", "request_duration_seconds"):  sc.requestTime,
		prometheus.BuildFQName(PrometheusNamespace, "", "request_size"):              sc.requestLength,
		prometheus.BuildFQName(PrometheusNamespace, "", "response_duration_seconds"): sc.responseTime,
		prometheus.BuildFQName(PrometheusNamespace, "", "response_size"):             sc.responseLength,
		prometheus.BuildFQName(PrometheusNamespace, "", "bytes_sent"):                sc.bytesSent,
		prometheus.BuildFQName(PrometheusNamespace, "", "upstream_latency_seconds"):  sc.upstreamLatency,
	}

	return sc, nil
}

func (sc *SocketCollector) handleMessage(msg []byte) {
	klog.V(5).InfoS("Metric", "message", string(msg))
	// klog.InfoS("Metric", "message", string(msg))

	// Unmarshal bytes
	var statsBatch []socketData
	err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(msg, &statsBatch)
	if err != nil {
		klog.ErrorS(err, "Unexpected error deserializing JSON", "payload", string(msg))
		return
	}

	for _, stats := range statsBatch {
		// Note these must match the order in requestTags at the top
		requestLabels := prometheus.Labels{
			"status":  stats.Status,
			// "upstream_status":       stats.upstream.Status,
			"upstream_cache_status": stats.upstream.CacheStatus,
			"method":  stats.Method,
			"path":    stats.Path,
			"service": stats.Service,
		}
		collectorLabels := prometheus.Labels{
			"status":  stats.Status,
			"service": stats.Service,
		}
		latencyLabels := prometheus.Labels{
			"service": stats.Service,
		}
		requestsMetric, err := sc.requests.GetMetricWith(collectorLabels)
		if err != nil {
			klog.ErrorS(err, "Error fetching requests metric")
		} else {
			requestsMetric.Inc()
		}
		if stats.Latency != -1 {
			latencyMetric, err := sc.upstreamLatency.GetMetricWith(latencyLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching latency metric")
			} else {
				latencyMetric.Observe(stats.Latency)
			}
		}

		if stats.RequestTime != -1 {
			requestTimeMetric, err := sc.requestTime.GetMetricWith(requestLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching request duration metric")
			} else {
				requestTimeMetric.Observe(stats.RequestTime)
			}
		}

		if stats.RequestLength != -1 {
			requestLengthMetric, err := sc.requestLength.GetMetricWith(requestLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching request length metric")
			} else {
				requestLengthMetric.Observe(stats.RequestLength)
			}
		}

		if stats.ResponseTime != -1 {
			responseTimeMetric, err := sc.responseTime.GetMetricWith(requestLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching upstream response time metric")
			} else {
				responseTimeMetric.Observe(stats.ResponseTime)
			}
		}

		if stats.ResponseLength != -1 {
			bytesSentMetric, err := sc.bytesSent.GetMetricWith(requestLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching bytes sent metric")
			} else {
				bytesSentMetric.Observe(stats.ResponseLength)
			}

			responseSizeMetric, err := sc.responseLength.GetMetricWith(requestLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching bytes sent metric")
			} else {
				responseSizeMetric.Observe(stats.ResponseLength)
			}
		}
	}
}

// Start listen for connections in the unix socket and spawns a goroutine to process the content
func (sc *SocketCollector) Start() {
	klog.InfoS("starting msg handling loop")
	for {
		conn, err := sc.listener.Accept()
		if err != nil {
			continue
		}
		go handleMessages(conn, sc.handleMessage)
	}
}

// Stop stops unix listener
func (sc *SocketCollector) Stop() {
	sc.listener.Close()
}

// Describe implements prometheus.Collector
func (sc SocketCollector) Describe(ch chan<- *prometheus.Desc) {
	sc.requestTime.Describe(ch)
	sc.requestLength.Describe(ch)
	sc.requests.Describe(ch)
	sc.upstreamLatency.Describe(ch)
	sc.responseTime.Describe(ch)
	sc.responseLength.Describe(ch)
	sc.bytesSent.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (sc SocketCollector) Collect(ch chan<- prometheus.Metric) {
	sc.requestTime.Collect(ch)
	sc.requestLength.Collect(ch)
	sc.requests.Collect(ch)
	sc.upstreamLatency.Collect(ch)
	sc.responseTime.Collect(ch)
	sc.responseLength.Collect(ch)
	sc.bytesSent.Collect(ch)
}

// handleMessages process the content received in a network connection
func handleMessages(conn io.ReadCloser, fn func([]byte)) {
	defer conn.Close()
	data, err := io.ReadAll(conn)
	if err != nil {
		return
	}

	fn(data)
}
