package agent

import (
	"log"
	"workshop3_dev/internals/models"
)

type OrchestratorFunc func(agent *Agent, job models.ServerResponse) models.AgentTaskResult

func registerCommands(agent *Agent) {
	// agent.commandOrchestrators["upload"] = (*Agent).orchestrateLoad
}

func (agent *Agent) ExecuteTask(job *models.ServerResponse) {
	log.Printf("AGENT IS NOW PROCESSING COMMAND %s with ID %s", job.Command, job.JobID)

}
