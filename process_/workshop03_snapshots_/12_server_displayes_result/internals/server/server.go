package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"math/rand"
	"net/http"
	"time"
	"workshop3_dev/internals/control"
	"workshop3_dev/internals/models"
)

// Server implements the Server interface for HTTPS
type Server struct {
	addr    string
	server  *http.Server
	tlsCert string
	tlsKey  string
}

// NewServer creates a new HTTPS server
func NewServer(addr string) *Server {
	return &Server{
		addr:    addr,
		tlsCert: "./certs/server.crt",
		tlsKey:  "./certs/server.key",
	}
}

// Start implements Server.Start for HTTPS
func (server *Server) Start() error {
	// Create Chi router
	r := chi.NewRouter()

	// Define our GET endpoint
	r.Get("/", RootHandler)

	// Define our POST endpoint for results
	r.Post("/results", ResultHandler)

	// Create the HTTP server
	server.server = &http.Server{
		Addr:    server.addr,
		Handler: r,
	}

	// Start the server
	return server.server.ListenAndServeTLS(server.tlsCert, server.tlsKey)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)

	var response models.ServerResponse

	// Check for pending commands
	cmd, exists := control.AgentCommands.GetCommand()
	if exists {
		log.Printf("Sending command to agent: %s\n", cmd.Command)
		response.Job = true
		response.Command = cmd.Command
		response.Arguments = cmd.Arguments
		response.JobID = fmt.Sprintf("job_%06d", rand.Intn(1000000))
		log.Printf("Job ID: %s\n", response.JobID)
	} else {
		log.Printf("No commands in queue")
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

// Stop implements Server.Stop for HTTPS
func (server *Server) Stop() error {
	// If there's no server, nothing to stop
	if server.server == nil {
		return nil
	}

	// Give the server 5 seconds to shut down gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.server.Shutdown(ctx)
}

// ResultHandler receives and displays the result from the Agent
func ResultHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)

	var result models.AgentTaskResult

	// Decode the incoming result
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		log.Printf("ERROR: Failed to decode JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error decoding JSON")
		return
	}

	// Unmarshal the CommandResult to get the actual message string
	var messageStr string
	if len(result.CommandResult) > 0 {
		if err := json.Unmarshal(result.CommandResult, &messageStr); err != nil {
			log.Printf("ERROR: Failed to unmarshal CommandResult: %v", err)
			messageStr = string(result.CommandResult) // Fallback to raw bytes as string
		}
	}

	if !result.Success {
		log.Printf("Job (ID: %s) has failed\nMessage: %s\nError: %v", result.JobID, messageStr, result.Error)
	} else {
		log.Printf("Job (ID: %s) has succeeded\nMessage: %s", result.JobID, messageStr)
	}
}
