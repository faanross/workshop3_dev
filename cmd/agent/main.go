package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
	"workshop3_dev/internals/agent"
)

func main() {

	serverAddr := "0.0.0.0:8443"
	delay := 5 * time.Second
	jitter := 50

	// Create our Agent instance
	newAgent := agent.NewAgent(serverAddr)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start run loop in goroutine
	go func() {
		log.Printf("Starting Agent Run Loop")
		log.Printf("Delay: %v, Jitter: %d%%", delay, jitter)

		if err := agent.RunLoop(newAgent, ctx, delay, jitter); err != nil {
			log.Printf("Run loop error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	log.Println("Shutting down client...")
	cancel() // This will cause the run loop to exit
}
