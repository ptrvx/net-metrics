package main

import (
	"fmt"
	"os"

	"github.com/BGrewell/go-iperf"
)

const (
	// iperf_host = "server-service.default.svc.cluster.local"
	iperf_host = "localhost"
)

func main() {

	c := iperf.NewClient(iperf_host)
	c.Debug = true
	c.SetStreams(4)
	c.SetTimeSec(30)
	c.SetInterval(1)
	c.SetPort(5201)
	liveReports := c.SetModeLive()

	go func() {
		for report := range liveReports {
			fmt.Println(report.String())
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
