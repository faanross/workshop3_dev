package control

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func StartControlAPI() {
	// Create Chi router
	r := chi.NewRouter()

	// Define the POST endpoint
	r.Get("/dummy", dummyHandler)

	log.Println("Starting Control API on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Printf("Control API error: %v", err)
		}
	}()
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("dummyHandler called")

	response := "Dummy endpoint triggered"

	json.NewEncoder(w).Encode(response)
}
