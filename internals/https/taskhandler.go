package https

import (
	"encoding/json"
	"errors"
	"log"
)

type OrchestratorFunc func(agent *HTTPSAgent, command string) AgentTaskResult

type AgentTaskResult struct {
	Command string `json:"command"`
	Success bool   `json:"success"`
	TaskID  string `json:"task_id"`
	Error   error  `json:"error,omitempty"`
}

func (agent *HTTPSAgent) ExecuteTask(job HTTPSResponse) {
	log.Printf("AGENT IS NOW PROCESSING COMMAND %s with ID %s", job.Command, job.JobID)

	var result AgentTaskResult

	orchestrator, found := agent.commandOrchestrators[job.Command]

	if found {
		result = orchestrator(agent, job.Command)
	} else {
		log.Printf("|WARN AGENT TASK| Received unknown command: '%s' (ID: %s)", job.Command, job.JobID)
		result = AgentTaskResult{
			Command: job.Command,
			Success: false,
			TaskID:  job.JobID,
			Error:   errors.New("command not found"),
		}
	}

	// Now marshall the result before sending it back
	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("|❗ERR AGENT TASK| Failed to marshal result for Task ID %s: %v", job.JobID, err)
		return // Cannot send result if marshalling fails
	}

	// Need to sent result back with new function

}

// orchestrateDownload is the orchestrator for the "download" command.
func (agent *HTTPSAgent) orchestrateLoad(task ServerTaskResponse) AgentTaskResult {

	// logic will go here

}
