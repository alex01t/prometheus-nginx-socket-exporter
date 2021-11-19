package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
)

// PrometheusNamespace default metric namespace
var PrometheusNamespace = "nginx"

func main() {
	sc, err := NewSocketCollector()
	if err != nil {
		fmt.Printf("failed to start nginx socker collector: %v\n", err)
		os.Exit(1)
	}
	go sc.Start()
	prometheus.MustRegister(sc)

	pc, err := NewNGINXProcess()
	if err != nil {
		fmt.Printf("failed to start nginx process collector: %v\n", err)
		os.Exit(1)
	}
	prometheus.MustRegister(pc)
	go pc.Start()

	nsc, err := NewNGINXStatus()
	if err != nil {
		fmt.Printf("failed to start nginx status collector: %v\n", err)
		os.Exit(1)
	}
	prometheus.MustRegister(nsc)
	go nsc.Start()

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Beginning to serve on port :10254")
	fmt.Println(http.ListenAndServe(":10254", nil))
}
