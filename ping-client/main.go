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
		panic(err)
	}

	metricsHandler := promhttp.HandlerFor(pingReg, promhttp.HandlerOpts{})

	http.Handle("/metrics", metricsHandler)
	go func() {
		fmt.Println("Starting metrics server...")
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

	pinger, err := probing.NewPinger(ipv4.String())
	if err != nil {
		panic(err)
	}
	fmt.Println("Pinger running...")
	pinger.Count = 3
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		err = pinger.Run() // Blocks until finished.
		if err != nil {
			panic(err)
		}
		stats := pinger.Statistics() // get send/receive/duplicate/rtt stats
		pingMetric.Set(float64(stats.AvgRtt))
	}

}
