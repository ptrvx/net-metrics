package main

import (
	"fmt"

	"github.com/BGrewell/go-iperf"
)

func main() {
	s := iperf.NewServer()
	s.SetPort(5201)
	s.Debug = true
	err := s.Start()
	if err != nil {
		fmt.Println("Failed to start server:", err)
		return
	}
	defer s.Stop()

	fmt.Println("Server is running...")
	select {} // Keeps the server running
}
