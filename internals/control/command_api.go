package control

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// CommandType represents a command with its arguments
type CommandType struct {
	Command   string          `json:"command"`
	Arguments json.RawMessage `json:"data,omitempty"`
}

type CommandValidator func(json.RawMessage) error

// Registry of valid commands with their validators and processors
// Registry of valid commands with their validators and processors
var validCommands = map[string]struct {
	Validator CommandValidator
	Processor CommandProcessor
}{
	"load": {
		Validator: validateLoadCommand,
		Processor: processLoadCommand,
	},
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
	cmdConfig, exists := validCommands[cmdType.Command]
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

	// Validate arguments
	if err := cmdConfig.Validator(cmdType.Arguments); err != nil {
		log.Printf("ERROR: Validation failed for '%s': %v", cmdType.Command, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  fmt.Sprintf("Validation failed: %v", err),
		})
		return
	}

	// Confirm on the client side command was received
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Received command")

}

// validateLoadCommand validates "load" command arguments from client
func validateLoadCommand(rawArgs json.RawMessage) error {
	if len(rawArgs) == 0 {
		return fmt.Errorf("load command requires arguments")
	}

	var args struct {
		FilePath   string `json:"file_path"`
		ExportName string `json:"export_name"`
	}

	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return fmt.Errorf("invalid argument format: %w", err)
	}

	if args.FilePath == "" {
		return fmt.Errorf("file_path is required")
	}

	if args.ExportName == "" {
		return fmt.Errorf("export_name is required")
	}

	// Check if file exists
	if _, err := os.Stat(args.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", args.FilePath)
	}

	log.Printf("Validation passed: file_path=%s, export_name=%s", args.FilePath, args.ExportName)

	return nil
}
