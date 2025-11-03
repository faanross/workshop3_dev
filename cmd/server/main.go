package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"workshop3_dev/internals/config"
	"workshop3_dev/internals/control"
	"workshop3_dev/internals/server"
)

const pathToYAML = "./configs/config.yaml"

func main() {
	// Command line flag for config file path
	configPath := flag.String("config", pathToYAML, "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Load our control API
	control.StartControlAPI()

	newServer := server.NewServer(cfg)

	// Start server in goroutine
	go func() {
		log.Printf("Starting  server on %s", cfg.ServerAddr)
		if err := newServer.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down server...")

	if err := newServer.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

}
