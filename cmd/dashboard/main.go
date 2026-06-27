package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloudprobe/internal/config"
	"cloudprobe/internal/web"
)

func main() {
	cfg := config.Load()

	server, err := web.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	server.Stop()
}
