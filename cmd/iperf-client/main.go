package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/BGrewell/go-iperf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// iperf_host = "server-service.default.svc.cluster.local"
	iperf_host  = "127.0.0.1"
	metricsPort = 9097
)

var (
	iperfMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "iperf_bandwidth",
		Help: "Network bandwidth client -> server, measured by iperf3, in bits/s",
	})
)

func main() {
	iperfReg := prometheus.NewRegistry()

	if err := iperfReg.Register(iperfMetric); err != nil {
		panic(err)
	}

	iperfHandler := promhttp.HandlerFor(iperfReg, promhttp.HandlerOpts{})

	http.Handle("/metrics", iperfHandler)
	go func() {
		fmt.Println("Starting metrics server...")
		err := http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil)
		if err != nil {
			panic(err)
		}
	}()

	c := iperf.NewClient(iperf_host)
	c.SetStreams(4)
	c.SetTimeSec(30)
	c.SetInterval(1)
	c.SetPort(5201)
	liveReports := c.SetModeLive()

	go func() {
		for report := range liveReports {
			// consider addding other metrics like congestion window
			// also consider collecting multiple values and calculating average
			iperfMetric.Set(report.BitsPerSecond)
		}
	}()

	err := c.Start()
	if err != nil {
		fmt.Printf("failed to start client: %v\n", err)
		os.Exit(-1)
	}

	fmt.Println("Watching live reports...")
	<-c.Done

	fmt.Println(c.Report().String())
}
