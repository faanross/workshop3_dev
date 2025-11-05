package control

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"workshop3_dev/internals/models"
)

func commandHandler(w http.ResponseWriter, r *http.Request) {

	// Instantiate custom type to receive command from client
	var cmdClient models.CommandClient

	// The first thing we need to do is unmarshall the request body into the custom type
	if err := json.NewDecoder(r.Body).Decode(&cmdClient); err != nil {
		log.Printf("ERROR: Failed to decode JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error decoding JSON")
		return
	}

	// Visually confirm we get the command we expected
	var commandReceived = fmt.Sprintf("Received command: %s", cmdClient.Command)
	log.Printf(commandReceived)

	// Check if command exists
	cmdConfig, exists := validCommands[cmdClient.Command] // Replace _ with cmdConfig
	if !exists {
		var commandInvalid = fmt.Sprintf("ERROR: Unknown command: %s", cmdClient.Command)
		log.Printf(commandInvalid)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(commandInvalid)
		return
	}

	// Validate arguments
	if err := cmdConfig.Validator(cmdClient.Arguments); err != nil {
		var commandInvalid = fmt.Sprintf("ERROR: Validation failed for '%s': %v", cmdClient.Command, err)
		log.Printf(commandInvalid)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(commandInvalid)
		return
	}

	// Process arguments (e.g., load file and convert to base64)
	processedArgs, err := cmdConfig.Processor(cmdClient.Arguments)
	if err != nil {
		var commandInvalid = fmt.Sprintf("ERROR: Processing failed for '%s': %v", cmdClient.Command, err)
		log.Printf(commandInvalid)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(commandInvalid)
		return

	}

	// Update command with processed arguments
	cmdClient.Arguments = processedArgs
	log.Printf("Processed command arguments: %s", cmdClient.Command)

	// Queue the validated and processed command
	AgentCommands.addCommand(cmdClient)

	// Confirm on the client side command was received
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(commandReceived)

}

func StartControlAPI() {
	// Create Chi router
	r := chi.NewRouter()

	// Define the POST endpoint
	r.Post("/command", commandHandler)

	log.Println("Starting Control API on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Printf("Control API error: %v", err)
		}
	}()
}

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

// CommandValidator validates command-specific arguments
type CommandValidator func(json.RawMessage) error

// CommandProcessor processes command-specific arguments (e.g., file loading)
type CommandProcessor func(json.RawMessage) (json.RawMessage, error)

// validateLoadCommand validates "load" command arguments from client
func validateLoadCommand(rawArgs json.RawMessage) error {
	if len(rawArgs) == 0 {
		return fmt.Errorf("load command requires arguments")
	}

	var args models.LoadArgsClient

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

	var clientArgs models.LoadArgsClient

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

	agentArgs := models.LoadArgsAgent{
		ShellcodeBase64: shellcodeB64,
		ExportName:      clientArgs.ExportName,
	}

	// Marshall arguments ready to be sent to agent
	processedJSON, err := json.Marshal(agentArgs)
	if err != nil {
		return nil, fmt.Errorf("marshaling processed args: %w", err)
	}

	log.Printf("Processed file: %s (%d bytes) -> base64 (%d chars)",
		clientArgs.FilePath, len(fileBytes), len(shellcodeB64))

	return processedJSON, nil
}

// CommandQueue stores commands ready for agent pickup
type CommandQueue struct {
	PendingCommands []models.CommandClient
	mu              sync.Mutex
}

// AgentCommands is Global command queue
var AgentCommands = CommandQueue{
	PendingCommands: make([]models.CommandClient, 0),
}

// addCommand adds a validated command to the queue
func (cq *CommandQueue) addCommand(command models.CommandClient) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	cq.PendingCommands = append(cq.PendingCommands, command)
	log.Printf("QUEUED: %s", command.Command)
}

// GetCommand retrieves and removes the next command from queue
func (cq *CommandQueue) GetCommand() (models.CommandClient, bool) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	if len(cq.PendingCommands) == 0 {
		return models.CommandClient{}, false
	}

	cmd := cq.PendingCommands[0]
	cq.PendingCommands = cq.PendingCommands[1:]

	log.Printf("DEQUEUED: Command '%s'", cmd.Command)

	return cmd, true
}
