package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloudprobe/pkg/agent/config"
	"cloudprobe/pkg/agent/reporter"
)

func main() {
	configPath := flag.String("c", "/etc/cloudprobe/agent.yml", "config file path")
	serverURL := flag.String("s", "", "dashboard server websocket url (overrides config)")
	token := flag.String("t", "", "agent token (overrides config)")
	flag.Parse()

	cfg := config.Load(*configPath)

	// 命令行参数优先
	if *serverURL != "" {
		cfg.ServerURL = *serverURL
	}
	if *token != "" {
		cfg.Token = *token
	}

	if cfg.ServerURL == "" {
		log.Fatal("Server URL is required. Use -s flag or config file.")
	}
	if cfg.Token == "" {
		log.Fatal("Agent token is required. Use -t flag or config file.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rep, err := reporter.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create reporter: %v", err)
	}

	go rep.Start(ctx)

	fmt.Println("CloudProbe Agent started.")
	fmt.Printf("Server: %s\n", cfg.ServerURL)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down agent...")
	cancel()
	rep.Stop()
}
