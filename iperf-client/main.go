package main

import (
	"fmt"
	"net"
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
		fmt.Printf("Starting /metrics server on port %v\n", metricsPort)
		err := http.ListenAndServe(fmt.Sprintf(":%v", metricsPort), nil)
		if err != nil {
			panic(err)
		}
	}()

	ips, err := net.LookupIP(host)
	if err != nil {
		panic(fmt.Sprintf("failed to lookup ip for %v: %v", host, err))
	}
	var ipv4 net.IP
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4 = ip
			break
		}
	}

	if ipv4 == nil {
		panic(fmt.Sprintf("failed to find IPv4 address for host: %v", host))
	}

	c := iperf.NewClient(ipv4.String())
	c.SetStreams(4)
	c.SetTimeSec(30)
	c.SetInterval(1)
	portInt, err := strconv.Atoi(port)
	if err != nil {
		panic(fmt.Sprintf("failed to parse port value %v: %v", port, err))
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
		panic(fmt.Sprintf("failed to start client: %v\n", err))
	}

	fmt.Printf("Watching live reports from %v:%v\n", c.Host(), c.Port())
	<-c.Done

	fmt.Println(c.Report().String())
}
