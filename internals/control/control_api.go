package control

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
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
}{
	"load": {
		Validator: validateLoadCommand,
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
