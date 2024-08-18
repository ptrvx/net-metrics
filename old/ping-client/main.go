package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	host        = os.Getenv("PING_TARGET")
	metricsPort = os.Getenv("METRICS_PORT")
)

var (
	pingMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ping_latency",
		Help: "Network ping latency, measured by pro-bing, in ns",
	})
)

func main() {
	pingReg := prometheus.NewRegistry()

	if err := pingReg.Register(pingMetric); err != nil {
		panic(fmt.Sprintf("failed to register metric: %v", err))
	}

	metricsHandler := promhttp.HandlerFor(pingReg, promhttp.HandlerOpts{})

	http.Handle("/metrics", metricsHandler)
	go func() {
		fmt.Printf("Starting metrics server on port: %v\n", metricsPort)
		err := http.ListenAndServe(fmt.Sprintf(":%v", metricsPort), nil)
		if err != nil {
			panic(fmt.Sprintf("failed to start metrics server: %v", err))
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

	pinger, err := probing.NewPinger(ipv4.String())
	if err != nil {
		panic(fmt.Sprintf("failed to start new pinger: %v", err))
	}
	fmt.Printf("Pinging host %v(%v)...", host, ipv4.String())
	pinger.Count = 3
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		err = pinger.Run() // Blocks until finished.
		if err != nil {
			// TODO: this doesn't need to panic
			fmt.Printf("failed to run ping: %v\n", err)
		}
		stats := pinger.Statistics() // get send/receive/duplicate/rtt stats
		pingMetric.Set(float64(stats.AvgRtt))
		fmt.Println(stats.AvgRtt)
	}

}
