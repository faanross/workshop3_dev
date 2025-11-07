package server

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"time"
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

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	response := "You have hit the server's root endpoint"

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
