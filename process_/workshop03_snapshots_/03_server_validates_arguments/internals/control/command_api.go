package control

import "encoding/json"

// Registry of valid commands with their validators and processors
var validCommands = map[string]struct {
	Validator CommandValidator
}{
	"shellcode": {
		Validator: validateShellcodeCommand,
	},
}

// CommandValidator validates command-specific arguments
type CommandValidator func(json.RawMessage) error
