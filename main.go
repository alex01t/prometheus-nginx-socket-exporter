package main

import (
  "fmt"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
  "net/http"
)

// PrometheusNamespace default metric namespace
var PrometheusNamespace = "nginx_socket"

func main() {
  sc, _ := NewSocketCollector("podname", "nsname", "classname", false)
  go sc.Start()
  prometheus.MustRegister(sc)

  http.Handle("/metrics", promhttp.Handler())
  fmt.Println("Beginning to serve on port :10254")
  fmt.Println(http.ListenAndServe(":10254", nil))
}
