package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"cloudprobe/pkg/agent"
	"cloudprobe/pkg/agent/config"
	"cloudprobe/pkg/agent/reporter"
)

func main() {
	configPath := flag.String("c", "/etc/cloudprobe/agent.yml", "config file path")
	serverURL := flag.String("s", "", "dashboard server url (overrides config)")
	token := flag.String("t", "", "agent token (overrides config)")
	mode := flag.String("m", "auto", "connection mode: ws | grpc | auto")
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

	// 自动检测模式
	connMode := *mode
	if connMode == "auto" {
		if strings.HasPrefix(cfg.ServerURL, "grpc://") || strings.HasPrefix(cfg.ServerURL, "grpcs://") {
			connMode = "grpc"
		} else {
			connMode = "ws"
		}
	}

	var rep interface {
		Start(context.Context)
		Stop()
	}
	var err error

	switch connMode {
	case "grpc":
		rep, err = agent.NewGRPCReporter(cfg)
		if err != nil {
			log.Fatalf("Failed to create gRPC reporter: %v", err)
		}
		fmt.Println("CloudProbe Agent started (gRPC mode).")
	default:
		rep, err = reporter.New(cfg)
		if err != nil {
			log.Fatalf("Failed to create reporter: %v", err)
		}
		fmt.Println("CloudProbe Agent started (WebSocket mode).")
	}

	fmt.Printf("Server: %s\n", cfg.ServerURL)

	go rep.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down agent...")
	cancel()
	rep.Stop()
}
