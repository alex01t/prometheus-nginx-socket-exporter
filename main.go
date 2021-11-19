package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"k8s.io/klog/v2"
)

// PrometheusNamespace default metric namespace
var PrometheusNamespace = "nginx"

func main() {
	klog.InfoS("Starting prometheus-nginx-exporter collectors...")
	sc, err := NewSocketCollector()
	if err != nil {
		klog.ErrorS(err, "failed to start nginx socker collector")
		os.Exit(1)
	}
	go sc.Start()
	prometheus.MustRegister(sc)

	pc, err := NewNGINXProcess()
	if err != nil {
		klog.ErrorS(err, "failed to start nginx process collector")
		os.Exit(1)
	}
	prometheus.MustRegister(pc)
	go pc.Start()

	nsc, err := NewNGINXStatus()
	if err != nil {
		klog.ErrorS(err, "failed to start nginx status collector")
		os.Exit(1)
	}
	prometheus.MustRegister(nsc)
	go nsc.Start()

	http.Handle("/metrics", promhttp.Handler())
	klog.InfoS("Listening on port 10254")
	klog.Error(http.ListenAndServe(":10254", nil))
}
