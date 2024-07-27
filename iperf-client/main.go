package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/BGrewell/go-iperf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	host        = os.Getenv("IPERF_SERVER_HOST")
	port        = os.Getenv("IPERF_SERVER_PORT")
	metricsPort = os.Getenv("METRICS_PORT")
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

	metricsHandler := promhttp.HandlerFor(iperfReg, promhttp.HandlerOpts{})

	http.Handle("/metrics", metricsHandler)
	go func() {
		fmt.Println("Starting metrics server...")
		err := http.ListenAndServe(fmt.Sprintf(":%v", metricsPort), nil)
		if err != nil {
			panic(err)
		}
	}()

	c := iperf.NewClient(host)
	c.SetStreams(4)
	c.SetTimeSec(30)
	c.SetInterval(1)
	portInt, err := strconv.Atoi(port)
	if err != nil {
		fmt.Printf("failed to parse port value %v: %v", port, err)
		os.Exit(-1)
	}
	c.SetPort(portInt)
	liveReports := c.SetModeLive()

	go func() {
		for report := range liveReports {
			// consider addding other metrics like congestion window
			// also consider collecting multiple values and calculating average
			iperfMetric.Set(report.BitsPerSecond)
		}
	}()

	err = c.Start()
	if err != nil {
		fmt.Printf("failed to start client: %v\n", err)
		os.Exit(-1)
	}

	fmt.Println("Watching live reports...")
	<-c.Done

	fmt.Println(c.Report().String())
}
