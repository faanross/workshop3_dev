package main

import (
	"log"
	"os"
	"os/signal"
	"workshop3_dev/internals/control"
	"workshop3_dev/internals/server"
)

func main() {

	serverInterface := "0.0.0.0:8443"

	// Load our control API
	control.StartControlAPI()

	newServer := server.NewServer(serverInterface)

	// Start server in goroutine
	go func() {
		log.Printf("Starting  server on %s", serverInterface)
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
