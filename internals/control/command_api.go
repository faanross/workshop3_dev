package control

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

type CommandType struct {
	Command   string          `json:"command"`
	Arguments json.RawMessage `json:"data,omitempty"`
}

type CommandQueue struct {
	// Queue of commands for any agent
	PendingCommands []string
	mu              sync.Mutex
}

var AgentCommands = CommandQueue{
	PendingCommands: make([]string, 0),
}

// Valid Commands
var validCommands = map[string]struct{}{
	"load": {},
	// Add other commands here
}

func handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the JSON body from the request
	var cmdType CommandType

	err := json.NewDecoder(r.Body).Decode(&cmdType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert received command to lowercase for comparison
	receivedCmd := strings.ToLower(cmdType.Command)

	// Set content-type for the response
	w.Header().Set("Content-Type", "application/json")
	var response string

	// Check if the lowercase command exists in the validCommands map
	if _, ok := validCommands[receivedCmd]; ok {
		// --- IF VALID ---
		log.Printf("VALID Command Received: %s", cmdType.Command)
		response = "Command is valid, accepted"

		// We need to send
		AgentCommands.addCommand(cmdType)
	} else {
		// --- IF INVALID ---
		log.Printf("INVALID Command Received: %s", cmdType.Command)
		response = "Command is invalid, rejected"
	}

	// Send the appropriate response
	json.NewEncoder(w).Encode(response)
}

func (cq *CommandQueue) addCommand(command CommandType) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	cq.PendingCommands = append(cq.PendingCommands, command.Command)

	log.Printf("QUEUEING: %s", command)
}

func (cq *CommandQueue) GetCommand() (string, bool) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	if len(cq.PendingCommands) == 0 {
		return "", false
	}

	// Get the first command in the queue
	cmd := cq.PendingCommands[0]

	// Remove it from the queue
	cq.PendingCommands = cq.PendingCommands[1:]

	log.Printf("Command retrieved: %s\n", cmd)

	return cmd, true
}
