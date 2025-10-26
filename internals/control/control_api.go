package control

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"sync"
)

// TransitionManager handles the global transition state
type TransitionManager struct {
	mu               sync.RWMutex
	shouldTransition bool
}

// Global instance
var Manager = &TransitionManager{
	shouldTransition: false,
}

// CheckAndReset atomically checks if transition is needed and resets the flag
// This ensures the transition signal is consumed only once
func (tm *TransitionManager) CheckAndReset() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.shouldTransition {
		tm.shouldTransition = false // Reset immediately
		log.Printf("Transition signal consumed and reset")
		return true
	}

	return false
}

// TriggerTransition sets the transition flag
func (tm *TransitionManager) TriggerTransition() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.shouldTransition = true
	log.Printf("Transition triggered")
}

func handleSwitch(w http.ResponseWriter, r *http.Request) {

	Manager.TriggerTransition()

	response := "Protocol transition triggered"

	json.NewEncoder(w).Encode(response)
}

func StartControlAPI() {
	// Create Chi router
	r := chi.NewRouter()

	// Define the POST endpoint
	r.Post("/switch", handleSwitch)

	log.Println("Starting Control API on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Printf("Control API error: %v", err)
		}
	}()
}
