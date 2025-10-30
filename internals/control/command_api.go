package control

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// CommandType represents a command with its arguments
type CommandType struct {
	Command   string          `json:"command"`
	Arguments json.RawMessage `json:"data,omitempty"`
}

// Registry of valid commands with their validators and processors
var validCommands = map[string]struct {
}{
	"load": {},
}

// HTTP handler for receiving commands from clients
func handleCommand(w http.ResponseWriter, r *http.Request) {
	// Instantiate empty struct to receive command we got from client
	var cmdType CommandType

	// Error, and return error to client, if we cannot deserialize
	if err := json.NewDecoder(r.Body).Decode(&cmdType); err != nil {
		log.Printf("ERROR: Failed to decode JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	// Visually confirm we get the command we expected
	cmdType.Command = strings.ToLower(cmdType.Command) // ADD THIS
	log.Printf("Received command: %s", cmdType.Command)

	// Check if command exists
	_, exists := validCommands[cmdType.Command]
	if !exists {
		log.Printf("ERROR: Unknown command: %s", cmdType.Command)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  fmt.Sprintf("Unknown command: %s", cmdType.Command),
		})
		return
	}
	log.Printf("A valid command was requested: %s", cmdType.Command)

	// Confirm on the client side command was received
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Received command")

}
