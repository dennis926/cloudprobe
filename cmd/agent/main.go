package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloudprobe/internal/agent/config"
	"cloudprobe/internal/agent/reporter"
)

func main() {
	configPath := flag.String("c", "/etc/cloudprobe/agent.yml", "config file path")
	flag.Parse()

	cfg := config.Load(*configPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rep, err := reporter.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create reporter: %v", err)
	}

	go rep.Start(ctx)

	fmt.Println("CloudProbe Agent started.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down agent...")
	cancel()
	rep.Stop()
}
