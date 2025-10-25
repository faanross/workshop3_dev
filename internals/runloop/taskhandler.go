package runloop

import (
	"akkeDNSII/internals/https"
	"log"
)

func executeTask(job https.HTTPSResponse) {
	log.Printf("AGENT IS NOW PROCESSING JOB: %s", job.Command)

	// Instantiate GENERAL RESPONSE AND CMD-SPECIFIC RESPONSE
	// TODO -> define these

	// call the specific function responsible for executing specific command (ORCHESTRATOR)
	// Marshall result into general response struct

	// Send result back to server (new function)
	// TODO - server needs endpoint to receive + process results
}
