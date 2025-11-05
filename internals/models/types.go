package models

import "encoding/json"

// CommandClient represents a command with its arguments as sent by Client
type CommandClient struct {
	Command   string          `json:"command"`
	Arguments json.RawMessage `json:"data,omitempty"`
}

// LoadArgsClient contains the command-specific arguments for Load as sent by Client
type LoadArgsClient struct {
	FilePath   string `json:"file_path"`
	ExportName string `json:"export_name"`
}

// LoadArgsAgent contains the command-specific arguments for Load as sent to the Agent
type LoadArgsAgent struct {
	ShellcodeBase64 string `json:"shellcode_base64"`
	ExportName      string `json:"export_name"`
}
