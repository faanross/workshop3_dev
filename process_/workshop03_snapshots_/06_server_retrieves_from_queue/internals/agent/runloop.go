package agent

import (
	"context"
	"log"
	"math/rand"
	"time"
)

func RunLoop(agent *Agent, ctx context.Context, delay time.Duration, jitter int) error {

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			log.Println("Run loop cancelled")
			return nil
		default:
		}

		response, err := agent.Send(ctx)
		if err != nil {
			log.Printf("Error sending request: %v", err)
			// Don't exit - just sleep and try again
			time.Sleep(delay)
			continue // Skip to next iteration
		}

		if response.Job {
			log.Printf("Job received from Server\n-> Command: %s\n-> JobID: %s", response.Command, response.JobID)
		} else {
			log.Printf("No job from Server")
		}

		// Calculate sleep duration with jitter
		sleepDuration := CalculateSleepDuration(delay, jitter)

		log.Printf("Sleeping for %v", sleepDuration)

		// Sleep with cancellation support
		select {
		case <-time.After(sleepDuration):
			// Continue to next iteration
		case <-ctx.Done():

			log.Println("Run loop cancelled")
			return nil
		}
	}
}

// CalculateSleepDuration calculates the actual sleep time with jitter
func CalculateSleepDuration(baseDelay time.Duration, jitterPercent int) time.Duration {
	if jitterPercent == 0 {
		return baseDelay
	}

	// Calculate jitter range
	jitterRange := float64(baseDelay) * float64(jitterPercent) / 100.0

	// Random value between -jitterRange and +jitterRange
	jitter := (rand.Float64()*2 - 1) * jitterRange

	// Calculate final duration
	finalDuration := float64(baseDelay) + jitter

	// Ensure we don't go negative
	if finalDuration < 0 {
		finalDuration = 0
	}

	return time.Duration(finalDuration)
}
